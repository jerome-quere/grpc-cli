package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
)

type DialConfig struct {
	Target    string
	Timeout   time.Duration
	TLSConfig *tls.Config
}

func dial(ctx context.Context, config *DialConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	if config.TLSConfig == nil {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config.TLSConfig)))
	}

	conn, err := grpc.DialContext(
		ctx,
		config.Target,
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot dial %s: %s", config.Target, err)
	}

	return conn, nil
}
