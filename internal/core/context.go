package core

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextDataKey struct{}

var _contextDataKey = contextDataKey{}

type contextData struct {
	BinaryName string
	Stderr     io.Writer
	Stdout     io.Writer
	Stdin      io.Reader
	Logger     *logrus.Logger

	DialConfig *DialConfig
	Connection *grpc.ClientConn
	MD         metadata.MD
}

func ctxInjectData(ctx context.Context, data *contextData) context.Context {
	return context.WithValue(ctx, _contextDataKey, data)
}

func ctxData(ctx context.Context) *contextData {
	return ctx.Value(_contextDataKey).(*contextData)
}

func CtxStderr(ctx context.Context) io.Writer {
	return ctxData(ctx).Stderr
}

func CtxStdout(ctx context.Context) io.Writer {
	return ctxData(ctx).Stdout
}

func CtxBinaryName(ctx context.Context) string {
	return ctxData(ctx).BinaryName
}

func CtxMD(ctx context.Context) metadata.MD {
	return ctxData(ctx).MD
}

func CtxLogger(ctx context.Context) *logrus.Logger {
	return ctxData(ctx).Logger
}

func CtxGrpcConnection(ctx context.Context) (*grpc.ClientConn, error) {
	data := ctxData(ctx)
	if data.Connection != nil {
		return data.Connection, nil
	}

	conn, err := dial(ctx, data.DialConfig)
	data.Connection = conn
	return data.Connection, err
}
