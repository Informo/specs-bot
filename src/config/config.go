package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the top-level structure of the configuration file.
type Config struct {
	Matrix  MatrixConfig  `yaml:"matrix"`
	Webhook WebhookConfig `yaml:"webhook"`
	Notices NoticesConfig `yaml:"notices"`
}

// MatrixConfig represents the Matrix part of the configuration file.
type MatrixConfig struct {
	HSURL       string `yaml:"hs_url"`
	MXID        string `yaml:"mxid"`
	AccessToken string `yaml:"access_token"`
}

// WebhookConfig represents the webhook part of the configuration file.
type WebhookConfig struct {
	Path       string `yaml:"path"`
	Secret     string `yaml:"secret"`
	ListenAddr string `yaml:"listen_addr"`
}

// NoticesConfig represents the notices part of the configurations file. It
// also contains a map of strings that will be filled from the strings JSON
// file.
type NoticesConfig struct {
	Pattern         string `yaml:"pattern"`
	Room            string `yaml:"room"`
	StringsFilePath string `yaml:"strings_file"`
	Strings         map[string]map[string]string
}

// Load reads the configuration file located at the provided path, and fills the
// properties of an instance of the Config structure with its content. It also
// loads the strings from the notices strings JSON file by parsing it.
func Load(filePath string) (cfg *Config, err error) {
	cfg = new(Config)

	rawCfg, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(rawCfg, cfg); err != nil {
		return
	}

	strings, err := ioutil.ReadFile(cfg.Notices.StringsFilePath)
	if err != nil {
		return
	}

	err = json.Unmarshal(strings, &(cfg.Notices.Strings))

	return
}
