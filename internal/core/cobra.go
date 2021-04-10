package core

import (
	"context"
	"fmt"

	"github.com/jerome-quere/grpc-cli/internal/args"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func buildRootCommand(ctx context.Context, files *protoregistry.Files) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: CtxBinaryName(ctx),

		// Do not display error with cobra, we handle it in bootstrap.
		SilenceErrors: true,

		// Do not display usage on error.
		SilenceUsage: true,
	}
	rootCmd.SetOut(CtxStderr(ctx))
	rootCmd.AddCommand(AutocompleteCommand(ctx, files))

	rpcCmd := &cobra.Command{
		Use:   "rpc",
		Short: "Execute an rpc call",
	}
	rootCmd.AddCommand(rpcCmd)

	files.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		for i := 0; i < file.Services().Len(); i++ {
			service := file.Services().Get(i)
			serviceCmd := &cobra.Command{
				Use: string(service.FullName()),
			}

			for i := 0; i < service.Methods().Len(); i++ {
				method := service.Methods().Get(i)
				serviceCmd.AddCommand(&cobra.Command{
					Use:  string(method.Name()),
					RunE: defaultRun(ctx, method),
				})
			}

			rpcCmd.AddCommand(serviceCmd)
		}
		return true
	})

	return rootCmd, nil
}

func defaultRun(ctx context.Context, method protoreflect.MethodDescriptor) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, rawArgs []string) error {

		req := dynamicpb.NewMessage(method.Input())
		res := dynamicpb.NewMessage(method.Output())

		err := args.Unmarshal(rawArgs, req)
		if err != nil {
			return fmt.Errorf("cannot unmarshal args: %s", err)
		}

		conn, err := CtxGrpcConnection(ctx)
		if err != nil {
			return fmt.Errorf("cannot get grpc connection: %s", err)
		}

		ctx := metadata.NewOutgoingContext(ctx, CtxMD(ctx))

		service := method.Parent().(protoreflect.ServiceDescriptor)
		method := fmt.Sprintf("/%s/%s", service.FullName(), method.Name())
		err = conn.Invoke(ctx, method, req, res)
		if err != nil {
			return fmt.Errorf("error while infoking rpc: %s", err)
		}

		raw, err := protojson.MarshalOptions{
			Multiline:       true,
			Indent:          "  ",
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(req)

		if err != nil {
			return fmt.Errorf("cannont marshal response: %s", err)
		}

		_, err = CtxStdout(ctx).Write(raw)
		if err != nil {
			return fmt.Errorf("cannont write response: %s", err)
		}

		return nil
	}
}
