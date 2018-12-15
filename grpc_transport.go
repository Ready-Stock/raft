package raft

import (
	"context"
	"google.golang.org/grpc"
	"io"
	"sync"
)

func NewGrpcTransport(server *grpc.Server, localAddr string) (*GrpcTransport, error) {
	return newGrpcTransport(server, localAddr)
}

type WithProtocolVersion interface {
	GetProtocolVersion() ProtocolVersion
}

type GrpcTransport struct {
	connPool        lake
	consumeChan     chan RPC
	heartbeatFn     func(RPC)
	heartbeatFnLock sync.Mutex

	shutdown     bool
	shutdownChan chan struct{}
	shutdownLock sync.Mutex

	gRPC      *grpc.Server
	svc       service
	localAddr string
}

func newGrpcTransport(server *grpc.Server, localAddr string) (*GrpcTransport, error) {
	transport := &GrpcTransport{
		connPool:     newLake(),
		consumeChan:  make(chan RPC),
		shutdownChan: make(chan struct{}),
		gRPC:         server,
		localAddr:    localAddr,
	}
	transport.svc = service{
		appendEntries:   transport.handleAppendEntriesCommand,
		requestVote:     transport.handleRequestVoteCommand,
		installSnapshot: transport.handleInstallSnapshotCommand,
	}
	RegisterRaftServiceServer(server, transport.svc)
	return transport, nil
}

func (transport *GrpcTransport) Consumer() <-chan RPC {
	return transport.consumeChan
}

func (transport *GrpcTransport) LocalAddr() ServerAddress {
	return ServerAddress(transport.localAddr)
}

func (transport *GrpcTransport) AppendEntriesPipeline(id ServerID, target ServerAddress) (AppendPipeline, error) {

	return nil, nil
}

func (transport *GrpcTransport) AppendEntries(id ServerID, target ServerAddress, args *AppendEntriesRequest, resp *AppendEntriesResponse) error {
	result, err := transport.executeTransportClient(context.Background(), id, target, func(ctx context.Context, client RaftServiceClient) (result interface{}, err error) {
		return client.AppendEntries(ctx, args)
	})
	*resp = *result.(*AppendEntriesResponse)
	return err
}

func (transport *GrpcTransport) RequestVote(id ServerID, target ServerAddress, args *RequestVoteRequest, resp *RequestVoteResponse) error {
	result, err := transport.executeTransportClient(context.Background(), id, target, func(ctx context.Context, client RaftServiceClient) (result interface{}, err error) {
		return client.RequestVote(ctx, args)
	})
	*resp = *result.(*RequestVoteResponse)
	return err
}

func (transport *GrpcTransport) InstallSnapshot(id ServerID, target ServerAddress, args *InstallSnapshotRequest, resp *InstallSnapshotResponse, data io.Reader) error {
	return nil
}

func (transport *GrpcTransport) EncodePeer(id ServerID, addr ServerAddress) []byte {
	return nil
}

func (transport *GrpcTransport) DecodePeer(data []byte) ServerAddress {
	return ""
}

func (transport *GrpcTransport) SetHeartbeatHandler(cb func(rpc RPC)) {
	transport.heartbeatFnLock.Lock()
	defer transport.heartbeatFnLock.Unlock()
	transport.heartbeatFn = cb
}

// Implementing the WithClose interface.
func (transport *GrpcTransport) Close() error {
	transport.shutdownLock.Lock()
	defer transport.shutdownLock.Unlock()

	if !transport.shutdown {
		close(transport.shutdownChan)
		transport.shutdown = true
	}
	return nil
}

func (transport *GrpcTransport) executeTransportClient(
	ctx context.Context,
	id ServerID,
	target ServerAddress,
	call func(ctx context.Context, client RaftServiceClient) (interface{}, error)) (result interface{}, err error) {
	conn, err := transport.connPool.GetConnection(ctx, target)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := NewRaftServiceClient(conn.ClientConn)
	return call(ctx, client)
}

func (transport *GrpcTransport) listen() {
	RegisterRaftServiceServer(transport.gRPC, transport.svc)
}

func (transport *GrpcTransport) handleAppendEntriesCommand(ctx context.Context, request *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	// Create the RPC object
	respCh := make(chan RPCResponse, 1)
	rpc := RPC{
		RespChan: respCh,
	}

	// Decode the command
	isHeartbeat := false

	rpc.Command = &request

	// Check if this is a heartbeat
	if request.Term != 0 && request.Leader != nil &&
		request.PrevLogEntry == 0 && request.PrevLogTerm == 0 &&
		len(request.Entries) == 0 && request.LeaderCommitIndex == 0 {
		isHeartbeat = true
	}

	// Check for heartbeat fast-path
	if isHeartbeat {
		transport.heartbeatFnLock.Lock()
		fn := transport.heartbeatFn
		transport.heartbeatFnLock.Unlock()
		if fn != nil {
			fn(rpc)
			goto RESP
		}
	}

	// Dispatch the RPC
	select {
	case transport.consumeChan <- rpc:
	case <-transport.shutdownChan:
		return nil, ErrTransportShutdown
	}

	// Wait for response
RESP:
	select {
	case resp := <-respCh:
		// Send the error first
		if resp.Error != nil {
			return nil, resp.Error
		}
		rsp := (resp.Response).(*AppendEntriesResponse)
		return rsp, resp.Error

	case <-transport.shutdownChan:
		return nil, ErrTransportShutdown
	}
}

func (transport *GrpcTransport) handleRequestVoteCommand(ctx context.Context, request *RequestVoteRequest) (*RequestVoteResponse, error) {
	// Create the RPC object
	respCh := make(chan RPCResponse, 1)
	rpc := RPC{
		RespChan: respCh,
	}

	// Decode the command
	rpc.Command = request
	// Dispatch the RPC
	select {
	case transport.consumeChan <- rpc:
	case <-transport.shutdownChan:
		return nil, ErrTransportShutdown
	}

	select {
	case resp := <-respCh:
		// Send the error first
		if resp.Error != nil {
			return nil, resp.Error
		}
		rsp := (resp.Response).(*RequestVoteResponse)
		return rsp, resp.Error

	case <-transport.shutdownChan:
		return nil, ErrTransportShutdown
	}
}

func (transport *GrpcTransport) handleInstallSnapshotCommand(ctx context.Context, request *InstallSnapshotRequest) (*InstallSnapshotResponse, error) {
	return nil, nil
}
