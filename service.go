package raft

import (
	"context"
)

type service struct {
	appendEntries   func(ctx context.Context, request *AppendEntriesRequest) (*AppendEntriesResponse, error)
	requestVote     func(ctx context.Context, request *RequestVoteRequest) (*RequestVoteResponse, error)
	installSnapshot func(ctx context.Context, request *InstallSnapshotRequestWrapper) (*InstallSnapshotResponse, error)
}

func (svc service) AppendEntries(ctx context.Context, request *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	return svc.appendEntries(ctx, request)
}

func (svc service) RequestVote(ctx context.Context, request *RequestVoteRequest) (*RequestVoteResponse, error) {
	return svc.requestVote(ctx, request)
}

func (svc service) InstallSnapshot(ctx context.Context, request *InstallSnapshotRequestWrapper) (*InstallSnapshotResponse, error) {
	return svc.installSnapshot(ctx, request)
}
