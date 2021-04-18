package config

import (
	"fmt"
	"os"

	"github.com/jerome-quere/grpc-cli/internal/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Profile  `yaml:",inline"`
	Profiles map[string]Profile `yaml:"profiles"`
}

const (
	DefaultProfileName = "default"
	DefaultConfigPath  = "~/.config/grpc-cli/config.yaml"
)

func LoadProfile(configPath string, profileName string) (Profile, error) {
	var config Config

	configPath = util.ResolvePath(configPath)

	if f, err := os.Open(configPath); !os.IsNotExist(err) {
		if err != nil {
			return Profile{}, fmt.Errorf("cannot open config file %s: %s", configPath, err)
		}

		defer f.Close()
		err = yaml.NewDecoder(f).Decode(&config)
		if err != nil {
			return Profile{}, fmt.Errorf("cannot parse config file %s: %s", configPath, err)
		}
	}

	profile := config.Profile
	if profileName != DefaultProfileName {
		profile = profile.Merge(config.Profiles[profileName])
	}

	return profile, nil
}

func (p Profile) Merge(p2 Profile) Profile {
	newProfile := p

	if p2.Target != nil {
		newProfile.Target = p2.Target
	}
	if p2.Descriptor != nil {
		newProfile.Descriptor = p2.Descriptor
	}
	if p2.CaCert != nil {
		newProfile.CaCert = p2.CaCert
	}
	if p2.Cert != nil {
		newProfile.Cert = p2.Cert
	}
	if p2.Key != nil {
		newProfile.Key = p2.Key
	}
	for k, v := range p2.Metadata {
		if _, exist := newProfile.Metadata[k]; !exist {
			newProfile.Metadata[k] = v
		}
	}
	return newProfile
}
