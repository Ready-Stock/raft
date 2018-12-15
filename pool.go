package raft

import (
	"context"
	"github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type lake struct {
	connections     map[ServerAddress]*pool
	connectionsLock sync.Mutex
}

func newLake() lake {
	return lake{
		connections:     map[ServerAddress]*pool{},
		connectionsLock: sync.Mutex{},
	}
}

func (l *lake) GetConnection(ctx context.Context, address ServerAddress) (conn *grpcpool.ClientConn, err error) {
	l.connectionsLock.Lock()
	defer l.connectionsLock.Unlock()

	p, ok := l.connections[address]
	if !ok {
		// A connection to this address hasn't been established yet.
		p, err = l.newPool(address)
		if err != nil {
			return nil, err
		}

		l.connections[address] = p
	}

	return p.GetConn(ctx)
}

type pool struct {
	address ServerAddress
	pool    *grpcpool.Pool
	factory grpcpool.Factory
}

func (l *lake) newPool(address ServerAddress) (pl *pool, err error) {
	p := &pool{
		address: address,
		factory: func() (*grpc.ClientConn, error) {
			conn, err := grpc.Dial(string(address), grpc.WithInsecure())
			if err != nil {
				return nil, err
			}
			return conn, err
		},
	}
	p.pool, err = grpcpool.New(p.factory, 2, 5, time.Second)
	return p, err
}

func (p *pool) GetConn(ctx context.Context) (*grpcpool.ClientConn, error) {
	return p.pool.Get(ctx)
}
