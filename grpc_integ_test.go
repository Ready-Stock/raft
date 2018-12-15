package raft

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

type GrpcRaftEnv struct {
	dir      string
	conf     *Config
	fsm      *MockFSM
	store    *InmemStore
	snapshot *FileSnapshotStore
	trans    *GrpcTransport
	raft     *Raft
	logger   *log.Logger
}

// Release shuts down and cleans up any stored data, its not restartable after this
func (r *GrpcRaftEnv) Release() {
	r.Shutdown()
	os.RemoveAll(r.dir)
}

// Shutdown shuts down raft & transport, but keeps track of its data, its restartable
// after a Shutdown() by calling Start()
func (r *GrpcRaftEnv) Shutdown() {
	r.logger.Printf("[WARN] Shutdown node at %v", r.raft.localAddr)
	f := r.raft.Shutdown()
	if err := f.Error(); err != nil {
		panic(err)
	}
	r.trans.Close()
}

// Restart will start a raft node that was previously Shutdown()
func (r *GrpcRaftEnv) Restart(t *testing.T) {
	// Transport 1 is consumer
	server, lis := getGrpcListener()
	// Transport 1 is consumer
	trans, err := NewGrpcTransport(server, lis.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	go func() {
		err := server.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	r.trans = trans
	r.logger.Printf("[INFO] Starting node at %v", trans.LocalAddr())
	raft, err := NewRaft(r.conf, r.fsm, r.store, r.store, r.snapshot, r.trans)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	r.raft = raft
}

func MakeGrpcRaft(t *testing.T, conf *Config, bootstrap bool) *GrpcRaftEnv {
	// Set the config
	if conf == nil {
		conf = inmemConfig(t)
	}

	dir, err := ioutil.TempDir("", "raft")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}

	stable := NewInmemStore()

	snap, err := NewFileSnapshotStore(dir, 3, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	env := &GrpcRaftEnv{
		conf:     conf,
		dir:      dir,
		store:    stable,
		snapshot: snap,
		fsm:      &MockFSM{},
		logger:   log.New(&testLoggerAdapter{t: t}, "", log.Lmicroseconds),
	}

	// Transport 1 is consumer
	server, lis := getGrpcListener()
	// Transport 1 is consumer
	trans, err := NewGrpcTransport(server, lis.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	go func() {
		err := server.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	env.logger = log.New(os.Stdout, string(trans.LocalAddr())+" :", log.Lmicroseconds)
	env.trans = trans

	if bootstrap {
		var configuration Configuration
		configuration.Servers = append(configuration.Servers, Server{
			Suffrage: Voter,
			ID:       conf.LocalID,
			Address:  trans.LocalAddr(),
		})
		err = BootstrapCluster(conf, stable, stable, snap, trans, configuration)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
	}
	log.Printf("[INFO] Starting node at %v", trans.LocalAddr())
	conf.Logger = env.logger
	raft, err := NewRaft(conf, env.fsm, stable, stable, snap, trans)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	env.raft = raft
	return env
}

func GrpcWaitFor(env *GrpcRaftEnv, state RaftState) error {
	limit := time.Now().Add(200 * time.Millisecond)
	for env.raft.State() != state {
		if time.Now().Before(limit) {
			time.Sleep(10 * time.Millisecond)
		} else {
			return fmt.Errorf("failed to transition to state %v", state)
		}
	}
	return nil
}

func GrpcWaitForAny(state RaftState, envs []*GrpcRaftEnv) (*GrpcRaftEnv, error) {
	limit := time.Now().Add(200 * time.Millisecond)
CHECK:
	for _, env := range envs {
		if env.raft.State() == state {
			return env, nil
		}
	}
	if time.Now().Before(limit) {
		goto WAIT
	}
	return nil, fmt.Errorf("failed to find node in %v state", state)
WAIT:
	time.Sleep(10 * time.Millisecond)
	goto CHECK
}

func GrpcCheckConsistent(envs []*GrpcRaftEnv, t *testing.T) {
	limit := time.Now().Add(400 * time.Millisecond)
	first := envs[0]
	first.fsm.Lock()
	defer first.fsm.Unlock()
	var err error
CHECK:
	l1 := len(first.fsm.logs)
	for i := 1; i < len(envs); i++ {
		env := envs[i]
		env.fsm.Lock()
		l2 := len(env.fsm.logs)
		if l1 != l2 {
			err = fmt.Errorf("log length mismatch %d %d", l1, l2)
			env.fsm.Unlock()
			goto ERR
		}
		for idx, log := range first.fsm.logs {
			other := env.fsm.logs[idx]
			if bytes.Compare(log, other) != 0 {
				err = fmt.Errorf("log entry %d mismatch between %s/%s : '%s' / '%s'", idx, first.raft.localAddr, env.raft.localAddr, log, other)
				env.fsm.Unlock()
				goto ERR
			}
		}
		env.fsm.Unlock()
	}
	return
ERR:
	if time.Now().After(limit) {
		t.Fatalf("%v", err)
	}
	first.fsm.Unlock()
	time.Sleep(20 * time.Millisecond)
	first.fsm.Lock()
	goto CHECK
}

// Tests Raft by creating a cluster, growing it to 5 nodes while
// causing various stressful conditions
func GrpcRaft_Integ(t *testing.T) {
	conf := DefaultConfig()
	conf.LocalID = ServerID("first")
	conf.HeartbeatTimeout = 50 * time.Millisecond
	conf.ElectionTimeout = 50 * time.Millisecond
	conf.LeaderLeaseTimeout = 50 * time.Millisecond
	conf.CommitTimeout = 5 * time.Millisecond
	conf.SnapshotThreshold = 100
	conf.TrailingLogs = 10

	// Create a single node
	env1 := MakeGrpcRaft(t, conf, true)
	err := GrpcWaitFor(env1, Leader)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	totalApplied := 0
	applyAndWait := func(leader *GrpcRaftEnv, n, sz int) {
		// Do some commits
		var futures []ApplyFuture
		for i := 0; i < n; i++ {
			futures = append(futures, leader.raft.Apply(logBytes(i, sz), 0))
		}
		for _, f := range futures {
			NoErr(WaitFuture(f, t), t)
			leader.logger.Printf("[DEBUG] Applied at %d, size %d", f.Index(), sz)
		}
		totalApplied += n
	}
	// Do some commits
	applyAndWait(env1, 100, 10)

	// Do a snapshot
	err = WaitFuture(env1.raft.Snapshot(), t)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Join a few nodes!
	var envs []*GrpcRaftEnv
	for i := 0; i < 4; i++ {
		conf.LocalID = ServerID(fmt.Sprintf("next-batch-%d", i))
		env := MakeGrpcRaft(t, conf, false)
		addr := env.trans.LocalAddr()
		// Do a snapshot
		err = WaitFuture(env1.raft.AddVoter(conf.LocalID, addr, 0, 0), t)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		envs = append(envs, env)
	}

	// Wait for a leader
	leader, err := GrpcWaitForAny(Leader, append([]*GrpcRaftEnv{env1}, envs...))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Do some more commits
	applyAndWait(leader, 100, 10)

	// Snapshot the leader
	err = WaitFuture(leader.raft.Snapshot(), t)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	GrpcCheckConsistent(append([]*GrpcRaftEnv{env1}, envs...), t)

	// shutdown a follower
	disconnected := envs[len(envs)-1]
	disconnected.Shutdown()

	// Do some more commits [make sure the resulting snapshot will be a reasonable size]
	applyAndWait(leader, 100, 10000)

	// snapshot the leader [leaders log should be compacted past the disconnected follower log now]
	err = WaitFuture(leader.raft.Snapshot(), t)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Unfortunately we need to wait for the leader to start backing off RPCs to the down follower
	// such that when the follower comes back up it'll run an election before it gets an rpc from
	// the leader
	time.Sleep(time.Second * 5)

	// start the now out of date follower back up
	disconnected.Restart(t)

	// wait for it to get caught up
	timeout := time.Now().Add(time.Second * 10)
	for disconnected.raft.getLastApplied() < leader.raft.getLastApplied() {
		time.Sleep(time.Millisecond)
		if time.Now().After(timeout) {
			t.Fatalf("Gave up waiting for follower to get caught up to leader")
		}
	}

	GrpcCheckConsistent(append([]*GrpcRaftEnv{env1}, envs...), t)

	// Shoot two nodes in the head!
	rm1, rm2 := envs[0], envs[1]
	rm1.Release()
	rm2.Release()
	envs = envs[2:]
	time.Sleep(10 * time.Millisecond)

	// Wait for a leader
	leader, err = GrpcWaitForAny(Leader, append([]*GrpcRaftEnv{env1}, envs...))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Do some more commits
	applyAndWait(leader, 100, 10)

	// Join a few new nodes!
	for i := 0; i < 2; i++ {
		conf.LocalID = ServerID(fmt.Sprintf("final-batch-%d", i))
		env := MakeGrpcRaft(t, conf, false)
		addr := env.trans.LocalAddr()
		err = WaitFuture(env1.raft.AddVoter(conf.LocalID, addr, 0, 0), t)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		envs = append(envs, env)
	}

	// Wait for a leader
	leader, err = GrpcWaitForAny(Leader, append([]*GrpcRaftEnv{env1}, envs...))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Remove the old nodes
	NoErr(WaitFuture(leader.raft.RemoveServer(rm1.raft.localID, 0, 0), t), t)
	NoErr(WaitFuture(leader.raft.RemoveServer(rm2.raft.localID, 0, 0), t), t)

	// Shoot the leader
	env1.Release()
	time.Sleep(3 * conf.HeartbeatTimeout)

	// Wait for a leader
	leader, err = GrpcWaitForAny(Leader, envs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	allEnvs := append([]*GrpcRaftEnv{env1}, envs...)
	GrpcCheckConsistent(allEnvs, t)

	if len(env1.fsm.logs) != totalApplied {
		t.Fatalf("should apply %d logs! %d", totalApplied, len(env1.fsm.logs))
	}

	for _, e := range envs {
		e.Release()
	}
}
