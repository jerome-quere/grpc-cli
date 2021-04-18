package core

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/jerome-quere/grpc-cli/internal/args"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func buildCobraCommand(ctx context.Context, files *protoregistry.Files) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: CtxBinaryName(ctx),

		// Do not display error with cobra, we handle it in bootstrap.
		SilenceErrors: true,

		// Do not display usage on error.
		SilenceUsage: true,
	}
	rootCmd.SetOut(CtxStderr(ctx))
	rootCmd.AddCommand(AutocompleteCobraCommand(ctx, files))
	rootCmd.AddCommand(RpcCobraCommand(ctx, files))
	return rootCmd, nil
}

func RpcCobraCommand(ctx context.Context, files *protoregistry.Files) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "rpc",
		Short: "Execute an rpc call",
	}

	files.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		for i := 0; i < file.Services().Len(); i++ {
			service := file.Services().Get(i)

			serviceCmd := &cobra.Command{
				Use: string(service.FullName()),
			}

			for i := 0; i < service.Methods().Len(); i++ {
				method := service.Methods().Get(i)
				methodCmd := &cobra.Command{
					Use:  string(method.Name()),
					RunE: rpcRun(ctx, method),
				}
				methodCmd.SetUsageTemplate(usageTemplate)
				methodCmd.Annotations = make(map[string]string)
				methodCmd.Annotations["UsageArgs"] = buildUsageArgs(ctx, method.Input())

				serviceCmd.AddCommand(methodCmd)
			}

			cmd.AddCommand(serviceCmd)
		}
		return true
	})

	return cmd
}

func buildUsageArgs(ctx context.Context, message protoreflect.MessageDescriptor) string {
	var argsBuffer bytes.Buffer
	tw := tabwriter.NewWriter(&argsBuffer, 0, 0, 3, ' ', 0)

	for i := 0; i < message.Fields().Len(); i++ {
		field := message.Fields().Get(i)
		argSpecUsageLeftPart := string(field.Name())
		argSpecUsageRightPart := field.Kind().String()
		if enum := field.Enum(); enum != nil {
			names := []string(nil)
			for i := 0; i < enum.Values().Len(); i++ {
				names = append(names, string(enum.Values().Get(i).Name()))
			}
			argSpecUsageRightPart += "(" + strings.Join(names, ", ") + ")"
		}

		_, err := fmt.Fprintf(tw, "  %s\t%s\n", argSpecUsageLeftPart, argSpecUsageRightPart)
		if err != nil {
			CtxLogger(ctx).Errorf("cannot build usage arg for message %s: %s", message.FullName(), err)
			return ""
		}
	}
	tw.Flush()

	paramsStr := strings.TrimSuffix(argsBuffer.String(), "\n")

	return paramsStr
}

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .Annotations.UsageArgs}}

ARGS:
{{.Annotations.UsageArgs}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func rpcRun(ctx context.Context, method protoreflect.MethodDescriptor) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, rawArgs []string) error {

		// Create gRPC request and response message
		req := dynamicpb.NewMessage(method.Input())
		res := dynamicpb.NewMessage(method.Output())

		// Unmarshal argument inside the gRPC request message
		err := args.Unmarshal(rawArgs, req)
		if err != nil {
			return fmt.Errorf("cannot unmarshal args: %s", err)
		}

		// Get the gRPC connection from the context.
		// Connection will be automatically closed in the Bootstrap method.
		conn, err := CtxGrpcConnection(ctx)
		if err != nil {
			return fmt.Errorf("cannot get grpc connection: %s", err)
		}

		// Injecting metadata in outgoing context
		ctx := metadata.NewOutgoingContext(ctx, CtxMD(ctx))

		// Executing gRPC call
		service := method.Parent().(protoreflect.ServiceDescriptor)
		method := fmt.Sprintf("/%s/%s", service.FullName(), method.Name())
		err = conn.Invoke(ctx, method, req, res)
		if err != nil {
			return fmt.Errorf("error while infoking rpc: %s", err)
		}

		// Marshal response in json
		raw, err := protojson.MarshalOptions{
			Multiline:       true,
			Indent:          "  ",
			UseProtoNames:   true,
			EmitUnpopulated: true,
		}.Marshal(req)

		if err != nil {
			return fmt.Errorf("cannont marshal response: %s", err)
		}

		// Write response on stdout
		_, err = CtxStdout(ctx).Write(raw)
		if err != nil {
			return fmt.Errorf("cannont write response: %s", err)
		}

		return nil
	}
}
