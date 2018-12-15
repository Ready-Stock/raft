package raft

// WithRPCHeader is an interface that exposes the RPC header.
type WithRPCHeader interface {
	GetRPCHeader() *RPCHeader
}

func (r AppendEntriesRequest) GetRPCHeader() *RPCHeader {
	return r.Header
}

func (r RequestVoteRequest) GetRPCHeader() *RPCHeader {
	return r.Header
}

func (r InstallSnapshotRequest) GetRPCHeader() *RPCHeader {
	return r.Header
}
