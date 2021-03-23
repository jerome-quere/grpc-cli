package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type BootstrapConfig struct {
	Stderr io.Writer
	Stdout io.Writer
	Stdin  io.Reader
	Args   []string
}

type MetadataFlags metadata.MD

func (m *MetadataFlags) String() string {
	return ""
}

func (m *MetadataFlags) Set(s string) error {
	h, err := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(s + "\n\n"))).ReadMIMEHeader()
	if err != nil {
		return err
	}
	for k, v := range h {
		delete(h, k)
		h[strings.ToLower(k)] = v
	}
	*m = MetadataFlags(metadata.Join(metadata.MD(*m), metadata.MD(h)))
	return nil
}

func (m *MetadataFlags) Type() string {
	return "test"
}

func (m *MetadataFlags) MD() metadata.MD {
	return metadata.MD(*m)
}

const defaultDialTimeout = time.Second * 10

func Bootstrap(ctx context.Context, config *BootstrapConfig) int {

	var descriptorPath string
	var target string
	var mdFlag MetadataFlags

	flags := pflag.NewFlagSet(config.Args[0], pflag.ContinueOnError)
	flags.StringVarP(&descriptorPath, "descriptor", "d", "./descriptor.pb", "Path to the descriptor file")
	flags.StringVarP(&target, "target", "t", "127.0.0.1:8080", "The grpc connection target")
	flags.VarP(&mdFlag, "metadata", "m", "Metadata to attache to the request")

	// Ignore unknown flag
	flags.ParseErrorsWhitelist.UnknownFlags = true
	// Make sure usage is never print by the parse method. (It should only be print by cobra)
	flags.Usage = func() {}

	// We don't do any error validation as:
	// Additional flag can be added on a per-command basis inside cobra
	// parse would fail as these flag are not known at this time.
	_ = flags.Parse(config.Args)

	files, err := loadDescriptorFile(descriptorPath)
	if err != nil {
		_, _ = fmt.Fprintf(config.Stderr, "cannot load descriptor file: %s", err)
		return -1
	}

	ctxData := &contextData{
		BinaryName: config.Args[0],
		Stdin:      config.Stdin,
		Stdout:     config.Stdout,
		Stderr:     config.Stderr,
		MD:         mdFlag.MD(),
		DialConfig: &DialConfig{
			Target:  target,
			Timeout: defaultDialTimeout,
		},
	}
	ctx = ctxInjectData(ctx, ctxData)
	defer func() {
		if ctxData.Connection != nil {
			ctxData.Connection.Close()
		}
	}()

	rootCmd, err := buildRootCommand(ctx, files)
	if err != nil {
		_, _ = fmt.Fprintf(config.Stderr, "cannot build cobra command: %s", err)
		return -1
	}

	rootCmd.PersistentFlags().StringVarP(&target, "descriptor", "d", "./descriptor.pb", "Path to the descriptor file")
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "127.0.0.1:8080", "The grpc connection target")
	rootCmd.PersistentFlags().VarP(&mdFlag, "metadata", "m", "Metadata to attache to the request")

	rootCmd.SetArgs(config.Args[1:])
	err = rootCmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(config.Stderr, "Error when executing cmd: %s", err)
		return -1
	}

	return 0
}

func loadDescriptorFile(path string) (*protoregistry.Files, error) {

	descriptorRaw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannont open file %s: %w", path, err)
	}

	var fileDescSet descriptorpb.FileDescriptorSet
	err = proto.Unmarshal(descriptorRaw, &fileDescSet)
	if err != nil {
		return nil, fmt.Errorf("cannot parse descriptor set: %w", err)
	}

	files, err := protodesc.NewFiles(&fileDescSet)
	if err != nil {
		return nil, fmt.Errorf("\"cannot create proto reflect struct: %w", err)
	}
	return files, nil
}
