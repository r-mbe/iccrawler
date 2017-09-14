package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/juju/errors"
)

// type SourceConfig struct {
// 	Schema string   `toml:"schema"`
// 	Tables []string `toml:"tables"`
// }

// Config struct
type Config struct {
	Debug  bool   `toml:"debug"`
	CSVOut string `toml:"csv_out"`
}

// NewConfigWithFile get a new config instance
func NewConfigWithFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return newConfig(string(data))

}

func newConfig(data string) (*Config, error) {
	var c Config

	_, err := toml.Decode(data, &c)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &c, nil
}
