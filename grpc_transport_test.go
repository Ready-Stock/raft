package raft

import (
	"github.com/gogo/protobuf/proto"
	"github.com/kataras/golog"
	"google.golang.org/grpc"
	"net"
	"sync"
	"testing"
	"time"
)

func getGrpcListener() (*grpc.Server, net.Listener) {
	golog.SetLevel("debug")
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	return grpcServer, lis
}

func TestGrpcTransport_CloseStreams(t *testing.T) {
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1, lis1.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	go func() {
		err := server1.Serve(lis1)
		if err != nil {
			panic(err)
		}
	}()

	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := AppendEntriesRequest{
		Term:         10,
		Leader:       []byte("cartman"),
		PrevLogEntry: 100,
		PrevLogTerm:  4,
		Entries: []*Log{
			{
				Index: 101,
				Term:  4,
				Type:  LogNoop,
			},
		},
		LeaderCommitIndex: 90,
	}
	resp := AppendEntriesResponse{
		Term:    4,
		LastLog: 90,
		Success: true,
	}

	// Listen for a request
	go func() {
		for {
			select {
			case rpc := <-rpcCh:
				// Verify the command
				req := rpc.Command.(*AppendEntriesRequest)
				if !proto.Equal(req, &args) {
					t.Fatalf("command mismatch: %#v %#v", req, args)
				}

				rpc.Respond(&resp, nil)
			case <-time.After(200 * time.Millisecond):
				return
			}
		}
	}()
	server2, lis2 := getGrpcListener()
	// Transport 2 makes outbound request, 3 conn pool
	trans2, err := NewGrpcTransport(server2, lis2.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	for i := 0; i < 2; i++ {
		// Create wait group
		wg := &sync.WaitGroup{}
		wg.Add(5)

		appendFunc := func() {
			defer wg.Done()
			var out AppendEntriesResponse
			if err := trans2.AppendEntries("id1", ServerAddress(lis1.Addr().String()), &args, &out); err != nil {
				t.Fatalf("err: %v", err)
			}
			// Verify the response
			if !proto.Equal(&resp, &out) {
				t.Fatalf("command mismatch: %#v %#v", resp, out)
			}
		}

		// Try to do parallel appends, should stress the conn pool
		for i := 0; i < 5; i++ {
			go appendFunc()
		}

		// Wait for the routines to finish
		wg.Wait()

		// Check the conn pool size
		// addr := trans1.LocalAddr()
		// if len(trans2.connPool[addr]) != 3 {
		//     t.Fatalf("Expected 3 pooled conns!")
		// }
		//
		// if i == 0 {
		//     trans2.CloseStreams()
		//     if len(trans2.connPool[addr]) != 0 {
		//         t.Fatalf("Expected no pooled conns after closing streams!")
		//     }
		// }
	}
}

func TestGrpcTransport_StartStop(t *testing.T) {
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1, lis1.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	trans1.Close()
}

func TestGrpcTransport_Heartbeat_FastPath(t *testing.T) {
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1, lis1.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	go func() {
		err := server1.Serve(lis1)
		if err != nil {
			panic(err)
		}
	}()
	// Make the RPC request
	args := AppendEntriesRequest{
		Term:   10,
		Leader: []byte("cartman"),
	}
	resp := AppendEntriesResponse{
		Term:    4,
		LastLog: 90,
		Success: true,
	}

	invoked := false
	fastPath := func(rpc RPC) {
		// Verify the command
		req := rpc.Command.(*AppendEntriesRequest)
		if !proto.Equal(req, &args) {
			t.Fatalf("command mismatch: %#v %#v", req, args)
		}

		rpc.Respond(&resp, nil)
		invoked = true
	}
	trans1.SetHeartbeatHandler(fastPath)

	server2, lis2 := getGrpcListener()
	// Transport 1 is consumer
	trans2, err := NewGrpcTransport(server2, lis2.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	var out AppendEntriesResponse
	if err := trans2.AppendEntries("id1", trans1.LocalAddr(), &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if !proto.Equal(&resp, &out) {
		t.Fatalf("command mismatch: %#v %#v", resp, out)
	}

	// Ensure fast-path is used
	if !invoked {
		t.Fatalf("fast-path not used")
	}
}

func TestGrpcTransport_AppendEntries(t *testing.T) {
	// Transport 1 is consumer
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1, lis1.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	go func() {
		err := server1.Serve(lis1)
		if err != nil {
			panic(err)
		}
	}()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := AppendEntriesRequest{
		Term:         10,
		Leader:       []byte("cartman"),
		PrevLogEntry: 100,
		PrevLogTerm:  4,
		Entries: []*Log{
			&Log{
				Index: 101,
				Term:  4,
				Type:  LogNoop,
			},
		},
		LeaderCommitIndex: 90,
	}
	resp := AppendEntriesResponse{
		Term:    4,
		LastLog: 90,
		Success: true,
	}

	// Listen for a request
	go func() {
		select {
		case rpc := <-rpcCh:
			// Verify the command
			req := rpc.Command.(*AppendEntriesRequest)
			if !proto.Equal(req, &args) {
				t.Fatalf("command mismatch: %#v %#v", req, args)
			}

			rpc.Respond(&resp, nil)

		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	server2, lis2 := getGrpcListener()
	// Transport 2 makes outbound request
	trans2, err := NewGrpcTransport(server2, lis2.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()

	var out AppendEntriesResponse
	if err := trans2.AppendEntries("id1", trans1.LocalAddr(), &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if !proto.Equal(&resp, &out) {
		t.Fatalf("command mismatch: %#v %#v", resp, out)
	}
}

func TestGrpcTransport_RequestVote(t *testing.T) {
	// Transport 1 is consumer
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1, lis1.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	go func() {
		err := server1.Serve(lis1)
		if err != nil {
			panic(err)
		}
	}()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := RequestVoteRequest{
		Term:         20,
		Candidate:    []byte("butters"),
		LastLogIndex: 100,
		LastLogTerm:  19,
	}
	resp := RequestVoteResponse{
		Term:    100,
		Granted: false,
	}

	// Listen for a request
	go func() {
		select {
		case rpc := <-rpcCh:
			golog.Infof("Received request from channel")
			// Verify the command
			req := rpc.Command.(RequestVoteRequest)
			if !proto.Equal(&req, &args) {
				t.Fatalf("command mismatch: %#v %#v", req, args)
			}

			rpc.Respond(&resp, nil)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	server2, lis2 := getGrpcListener()
	// Transport 2 makes outbound request
	trans2, err := NewGrpcTransport(server2, lis2.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()
	var out RequestVoteResponse
	if err := trans2.RequestVote("id1", trans1.LocalAddr(), &args, &out); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Verify the response
	if !proto.Equal(&resp, &out) {
		t.Fatalf("command mismatch: %#v %#v", resp, out)
	}
}

func TestGrpcTransport_AppendEntriesPipeline(t *testing.T) {
	// Transport 1 is consumer
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1, lis1.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans1.Close()
	go func() {
		err := server1.Serve(lis1)
		if err != nil {
			panic(err)
		}
	}()
	rpcCh := trans1.Consumer()

	// Make the RPC request
	args := AppendEntriesRequest{
		Term:         10,
		Leader:       []byte("cartman"),
		PrevLogEntry: 100,
		PrevLogTerm:  4,
		Entries: []*Log{
			&Log{
				Index: 101,
				Term:  4,
				Type:  LogNoop,
			},
		},
		LeaderCommitIndex: 90,
	}
	resp := AppendEntriesResponse{
		Term:    4,
		LastLog: 90,
		Success: true,
	}

	// Listen for a request
	go func() {
		for i := 0; i < 10; i++ {
			select {
			case rpc := <-rpcCh:
				// Verify the command
				req := rpc.Command.(*AppendEntriesRequest)
				if !proto.Equal(req, &args) {
					t.Fatalf("command mismatch: %#v %#v", req, args)
				}
				rpc.Respond(&resp, nil)

			case <-time.After(200 * time.Millisecond):
				t.Fatalf("timeout")
			}
		}
	}()

	server2, lis2 := getGrpcListener()
	// Transport 2 makes outbound request
	trans2, err := NewGrpcTransport(server2, lis2.Addr().String())
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer trans2.Close()
	pipeline, err := trans2.AppendEntriesPipeline("id1", trans1.LocalAddr())
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 10; i++ {
		out := new(AppendEntriesResponse)
		if _, err := pipeline.AppendEntries(&args, out); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	respCh := pipeline.Consumer()
	for i := 0; i < 10; i++ {
		select {
		case ready := <-respCh:
			// Verify the response
			if !proto.Equal(&resp, ready.Response()) {
				t.Fatalf("command mismatch: %#v %#v", &resp, ready.Response())
			}
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}
	pipeline.Close()
}
