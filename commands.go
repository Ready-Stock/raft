package raft

// WithRPCHeader is an interface that exposes the RPC header.
type WithRPCHeader interface {
	GetRPCHeader() *RPCHeader
}
