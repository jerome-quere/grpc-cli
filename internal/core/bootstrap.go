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
}

const defaultDialTimeout = time.Second * 10

func Bootstrap(ctx context.Context, bootstrapConfig *BootstrapConfig) int {

	logger := logrus.New()
	logger.SetFormatter(logrus.Formatter(&logrus.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	}))
	logger.SetOutput(bootstrapConfig.Stderr)

	flags := NewFlagSet(bootstrapConfig.Args[0])
	// We don't do any error validation as:
	// Additional flag can be added on a per-command basis inside cobra
	// parse would fail as these flag are not known at this time.
	_ = flags.Parse(bootstrapConfig.Args)

	logger.SetLevel(logrus.InfoLevel)
	if flags.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Load profile from config file and flags
	logger.Debugf("Loading config file from %s with profile %s", flags.Config, flags.Profile)
	profile, err := config.LoadProfile(flags.Config, flags.Profile)
	if err != nil {
		logger.Errorf("cannot load profile: %s", err)
		return 1
	}
	profile.Merge(flags.GetProfile())

	// Validate profile to make sure all required fields are set
	err = profile.Validate()
	if err != nil {
		logger.Errorf("error while validating profile: %s", err)
		return 1
	}
	logger.Debugf("Loading profile complete: %s", profile)

	// Loading proto FileSetDescriptor file
	logger.Debugf("Loading descriptor from file: %s", profile.GetDescriptor())
	files, err := loadDescriptorFile(profile.GetDescriptor())
	if err != nil {
		logger.Errorf("cannot load descriptor file: %s\n", err)
		return 1
	}
	logger.Debugf("Loading descriptor complete: %d proto files found", files.NumFiles())

	// Generating TLS configuration from the profile
	logger.Debugf("Loading tls config")
	tlsConfig, err := profile.GetTLSConfig()
	if err != nil {
		logger.Errorf("Cannont load TLS config: %s", err)
		return 1
	}
	logger.Debugf("Loading tls config complete")

	ctxData := &contextData{
		BinaryName: bootstrapConfig.Args[0],
		Stdin:      bootstrapConfig.Stdin,
		Stdout:     bootstrapConfig.Stdout,
		Stderr:     bootstrapConfig.Stderr,
		MD:         profile.Metadata,
		Logger:     logger,
		DialConfig: &DialConfig{
			TLSConfig: tlsConfig,
			Target:    profile.GetTarget(),
			Timeout:   defaultDialTimeout,
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
