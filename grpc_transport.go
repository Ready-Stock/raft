package raft

import (
	"context"
	"fmt"
	"github.com/kataras/go-errors"
	"google.golang.org/grpc"
	"io"
)

func NewGrpcTransport(server *grpc.Server) (*GrpcTransport, error) {
	return newGrpcTransport(server)
}

type WithProtocolVersion interface {
	GetProtocolVersion() ProtocolVersion
}

type GrpcTransport struct {
	connPool    lake
	consumeChan chan RPC
	gRPC        *grpc.Server
	svc         service
	localAddr   string
}

func newGrpcTransport(server *grpc.Server) (*GrpcTransport, error) {
	transport := &GrpcTransport{
		connPool:    newLake(),
		consumeChan: make(chan RPC, 0),
		gRPC:        server,
	}
	transport.svc = service{
		appendEntries:   transport.handleAppendEntriesCommand,
		requestVote:     transport.handleRequestVoteCommand,
		installSnapshot: transport.handleInstallSnapshotCommand,
	}
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

func (transport *GrpcTransport) AppendEntries(id ServerID, target ServerAddress, args *AppendEntriesRequest, resp *AppendEntriesResponse) (err error) {
	return transport.executeTransportClient(context.Background(), id, target, func(ctx context.Context, client RaftServiceClient) error {
		response, err := client.AppendEntries(ctx, args)
		if err != nil {
			return err
		}

		if !response.Success {
			return errors.New(fmt.Sprintf("append entries was not successful against node [%s] address [%s]", id, target))
		}
		return nil
	})
}

func (transport *GrpcTransport) RequestVote(id ServerID, target ServerAddress, args *RequestVoteRequest, resp *RequestVoteResponse) error {
	return transport.executeTransportClient(context.Background(), id, target, func(ctx context.Context, client RaftServiceClient) error {
		_, err := client.RequestVote(ctx, args)
		return err
	})
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

}

// Implementing the WithClose interface.
func (transport *GrpcTransport) Close() error {
	return nil
}

func (transport *GrpcTransport) executeTransportClient(ctx context.Context, id ServerID, target ServerAddress, call func(ctx context.Context, client RaftServiceClient) error) error {
	conn, err := transport.connPool.GetConnection(ctx, target)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := NewRaftServiceClient(conn.ClientConn)
	return call(ctx, client)
}

func (transport *GrpcTransport) listen() {
	RegisterRaftServiceServer(transport.gRPC, transport.svc)
}

func (transport *GrpcTransport) handleAppendEntriesCommand(ctx context.Context, request *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	return nil, nil
}

func (transport *GrpcTransport) handleRequestVoteCommand(ctx context.Context, request *RequestVoteRequest) (*RequestVoteResponse, error) {
	return nil, nil
}

func (transport *GrpcTransport) handleInstallSnapshotCommand(ctx context.Context, request *InstallSnapshotRequest) (*InstallSnapshotResponse, error) {
	return nil, nil
}
