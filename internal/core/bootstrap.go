package core

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/jerome-quere/grpc-cli/internal/config"
	"github.com/sirupsen/logrus"
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

	// Usefull fot test to provide a Descriptor directly.
	// If this is set this will bypass default Descriptor loading
	Descriptor []byte
}

const defaultDialTimeout = time.Second * 10

func Bootstrap(ctx context.Context, bootstrapConfig *BootstrapConfig) int {

	//
	// We do a first argument parsing before building cobra command because we need some argument like
	// config file or protoset file before building the cobra command tree
	//
	flags := NewFlagSet(bootstrapConfig.Args[0])
	// We don't do any error validation as flag parsing error will be handled in cobra Run later on
	_ = flags.Parse(bootstrapConfig.Args)

	//
	// Init Logger
	//
	logger := logrus.New()
	logger.SetFormatter(logrus.Formatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	}))
	logger.SetOutput(bootstrapConfig.Stderr)
	logger.SetLevel(logrus.InfoLevel)
	if flags.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	//
	// Load profile from config file and flags
	//
	logger.Debugf("Loading config file from %s with profile %s", flags.Config, flags.Profile)
	profile, err := config.LoadProfile(flags.Config, flags.Profile)
	if err != nil {
		logger.Errorf("cannot load profile: %s", err)
		return 1
	}
	// We merge profile params passed as flag on top of the config profile
	profile = profile.Merge(flags.GetProfile())

	// Validate profile to make sure all required fields are set
	err = profile.Validate()
	if err != nil {
		logger.Errorf("error while validating profile: %s", err)
		return 1
	}
	logger.Debugf("Loading profile complete: %s", profile)

	//
	// Loading proto descriptor file
	//
	var files *protoregistry.Files
	if bootstrapConfig.Descriptor == nil {
		logger.Debugf("Loading descriptor from file: %s", profile.GetDescriptor())
		files, err = loadDescriptorFromFile(profile.GetDescriptor())
		if err != nil {
			logger.Errorf("cannot load descriptor file: %s\n", err)
			return 1
		}
		logger.Debugf("Loading descriptor complete: %d proto files found", files.NumFiles())
	} else {
		logger.Debugf("Loading descriptor from bootstrap config")
		files, err = loadDescriptorFromBytes(bootstrapConfig.Descriptor)
		if err != nil {
			logger.Errorf("cannot load descriptor file: %s\n", err)
			return 1
		}
		logger.Debugf("Loading descriptor complete: %d proto files found", files.NumFiles())
	}

	//
	// Building gRPC dial config for the gRPC connection
	//
	logger.Debugf("Building dial config")
	tlsConfig, err := profile.GetTLSConfig()
	if err != nil {
		logger.Errorf("Cannont load TLS config: %s", err)
		return 1
	}
	dialConfig := &DialConfig{
		TLSConfig: tlsConfig,
		Target:    profile.GetTarget(),
		Timeout:   defaultDialTimeout,
	}

	logger.Debugf("Building dial config complete")

	//
	// Building context data. This data will be injected in context and
	// can be accessed using the various ctxExtract... functions
	//
	ctxData := &contextData{
		BinaryName: bootstrapConfig.Args[0],
		Stdin:      bootstrapConfig.Stdin,
		Stdout:     bootstrapConfig.Stdout,
		Stderr:     bootstrapConfig.Stderr,
		MD:         profile.Metadata,
		Logger:     logger,
		DialConfig: dialConfig,

		// We do not open connection now as we are not sure we need it yet.
		// gRPC connection will only be opened when required.
		Connection: nil,
	}
	ctx = ctxInjectData(ctx, ctxData)
	// If connection was opened during command execution we make sur to close it properly
	defer func() {
		if ctxData.Connection != nil {
			ctxData.Connection.Close()
		}
	}()

	// Build cobra command
	rootCmd, err := buildCobraCommand(ctx, files)
	if err != nil {
		logger.Errorf("cannot build cobra command: %s", err)
		return 1
	}

	rootCmd.PersistentFlags().AddFlagSet(flags.FlagSet)
	rootCmd.SetArgs(bootstrapConfig.Args[1:])
	err = rootCmd.Execute()
	if err != nil {
		logger.Errorf("error when executing cmd: %s\n", err)
		return 1
	}

	return 0
}

func loadDescriptorFromFile(path string) (*protoregistry.Files, error) {

	descriptorRaw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannont open file %s: %w", path, err)
	}
	return loadDescriptorFromBytes(descriptorRaw)
}

func loadDescriptorFromBytes(descriptorRaw []byte) (*protoregistry.Files, error) {
	var fileDescSet descriptorpb.FileDescriptorSet
	err := proto.Unmarshal(descriptorRaw, &fileDescSet)
	if err != nil {
		return nil, fmt.Errorf("cannot parse descriptor set: %w", err)
	}

	files, err := protodesc.NewFiles(&fileDescSet)
	if err != nil {
		return nil, fmt.Errorf("\"cannot create proto reflect struct: %w", err)
	}
	return files, nil
}
