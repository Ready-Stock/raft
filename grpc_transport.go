package raft

import "io"

type WithProtocolVersion interface {
	GetProtocolVersion() ProtocolVersion
}

type GrpcTransport struct {
	consumeChan chan RPC
}

func (transport *GrpcTransport) Consumer() <-chan RPC {
	return transport.consumeChan
}

func (transport *GrpcTransport) LocalAddr() ServerAddress {
	return ""
}

func (transport *GrpcTransport) AppendEntriesPipeline(id ServerID, target ServerAddress) (AppendPipeline, error) {
	return nil, nil
}

func (transport *GrpcTransport) AppendEntries(id ServerID, target ServerAddress, args *AppendEntriesRequest, resp *AppendEntriesResponse) error {
	return nil
}

func (transport *GrpcTransport) RequestVote(id ServerID, target ServerAddress, args *RequestVoteRequest, resp *RequestVoteResponse) error {
	return nil
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
