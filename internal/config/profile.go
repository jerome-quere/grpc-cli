package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/jerome-quere/grpc-cli/internal/util"
	"google.golang.org/grpc/metadata"
)

type Profile struct {
	Target     *string     `yaml:"target"`
	Descriptor *string     `yaml:"descriptor"`
	Metadata   metadata.MD `yaml:"metadata"`
	CaCert     *string     `yaml:"ca_cert"`
	Cert       *string     `yaml:"cert"`
	Key        *string     `yaml:"key"`
	DisableTLS *bool       `yaml:"disable_tls"`
}

func (p Profile) Validate() error {
	switch {
	case p.Target == nil:
		return fmt.Errorf("target cannot be empty, you must set it in the config file or pass it as argument")
	case p.Descriptor == nil:
		return fmt.Errorf("descriptor cannot be empty, you must set it in the config file or pass it as argument")
	default:
		return nil
	}
}

func (p Profile) GetDescriptor() string {
	return util.ResolvePath(*p.Descriptor)
}

func (p Profile) GetDisableTLS() bool {
	return p.DisableTLS != nil && *p.DisableTLS
}

func (p Profile) GetTLSConfig() (*tls.Config, error) {
	if p.GetDisableTLS() {
		return nil, nil
	}

	config := &tls.Config{}

	if p.CaCert != nil {
		caPem, err := ioutil.ReadFile(util.ResolvePath(*p.CaCert))
		if err != nil {
			return nil, fmt.Errorf("cannot read ca file: %s", err)
		}
		config.RootCAs = &x509.CertPool{}
		config.RootCAs.AppendCertsFromPEM(caPem)
	}

	if p.Cert != nil && p.Key != nil {
		cert, err := tls.LoadX509KeyPair(util.ResolvePath(*p.Cert), util.ResolvePath(*p.Key))
		if err != nil {
			return nil, fmt.Errorf("cannot read ca file: %s", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	return config, nil
}

func (p Profile) GetTarget() string {
	return *p.Target
}

func (p Profile) String() string {
	res, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	return string(res)
}
