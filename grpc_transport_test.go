package raft

import (
	"google.golang.org/grpc"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

func getGrpcListener() (*grpc.Server, net.Listener) {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	return grpcServer, lis
}

func GrpcTransport_CloseStreams(t *testing.T) {
	server1, lis1 := getGrpcListener()
	// Transport 1 is consumer
	trans1, err := NewGrpcTransport(server1)
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
		for {
			select {
			case rpc := <-rpcCh:
				// Verify the command
				req := rpc.Command.(*AppendEntriesRequest)
				if !reflect.DeepEqual(req, &args) {
					t.Fatalf("command mismatch: %#v %#v", *req, args)
				}
				rpc.Respond(&resp, nil)

			case <-time.After(200 * time.Millisecond):
				return
			}
		}
	}()
	server2, _ := getGrpcListener()
	// Transport 2 makes outbound request, 3 conn pool
	trans2, err := NewGrpcTransport(server2)
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
			if !reflect.DeepEqual(resp, out) {
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
		//addr := trans1.LocalAddr()
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
