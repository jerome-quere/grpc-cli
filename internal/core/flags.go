package core

import (
	"github.com/jerome-quere/grpc-cli/internal/config"
	"github.com/spf13/pflag"
)

type FlagSet struct {
	*pflag.FlagSet

	Config     string
	Profile    string
	Descriptor string
	Target     string
	Metadata   MetadataFlags
	CaCert     string
	Cert       string
	Key        string
	Verbose    bool
	DisableTLS bool
}

func NewFlagSet(binaryName string) *FlagSet {
	flags := &FlagSet{
		FlagSet: pflag.NewFlagSet(binaryName, pflag.ContinueOnError),
	}

	flags.StringVarP(&flags.Config, "config", "c", config.DefaultConfigPath, "Path to the config file")
	flags.StringVarP(&flags.Profile, "profile", "p", config.DefaultProfileName, "Config profile to load")
	flags.StringVarP(&flags.Descriptor, "descriptor", "d", "", "Path to the descriptor file")
	flags.StringVarP(&flags.Target, "target", "t", "", "The grpc connection target")
	flags.StringVarP(&flags.CaCert, "ca-cert", "", "", "Root CA you want to use. Use system by default")
	flags.StringVarP(&flags.Cert, "cert", "", "", "Client certificate path. (PEM format)")
	flags.StringVarP(&flags.Key, "key", "", "", "Client key path. (PEM format)")
	flags.BoolVarP(&flags.Verbose, "verbose", "v", false, "Enable verbose")
	flags.BoolVarP(&flags.DisableTLS, "disable-tls", "", false, "Enable verbose")
	flags.VarP(&flags.Metadata, "metadata", "m", "Metadata to attache to the request")

	flags.ParseErrorsWhitelist.UnknownFlags = true
	flags.Usage = func() {}
	return flags
}

func (fs *FlagSet) GetProfile() config.Profile {
	profile := config.Profile{
		Metadata: fs.Metadata.MD(),
	}
	if fs.Descriptor != "" {
		profile.Descriptor = &fs.Descriptor
	}
	if fs.Target != "" {
		profile.Target = &fs.Target
	}
	if fs.CaCert != "" {
		profile.CaCert = &fs.CaCert
	}
	if fs.Cert != "" {
		profile.Cert = &fs.Cert
	}
	if fs.Key != "" {
		profile.Key = &fs.Key
	}
	if fs.DisableTLS {
		profile.DisableTLS = &fs.DisableTLS
	}

	return profile
}
