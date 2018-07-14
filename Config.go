package manualNotify

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Zone struct {
}

type Config struct {
	Zones []struct {
		Name        string `yaml:"name"`
		Destination string `yaml:"destination"`
		IsSigned    bool   `yaml:"issigned"`
	} `yaml:"zones"`
	Hostname   string `yaml:"hostname"`
	Resolvconf string `yaml:"resolvconf"`
	Unitname   string `yaml:"unit"`
}

func LoadConfigFromFile(file string) (*Config, error) {
	cfg := Config{}

	if contents, err := ioutil.ReadFile(file); err == nil {
		err = yaml.Unmarshal(contents, &cfg)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &cfg, nil
}
