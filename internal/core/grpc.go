package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DialConfig struct {
	Target  string
	Timeout time.Duration
}

func dial(ctx context.Context, config *DialConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		config.Target,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot dial %s: %s", config.Target, err)
	}

	return conn, nil
}
