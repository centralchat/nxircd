package config

import (
	"encoding/json"
	"io/ioutil"
)

func IsEnabled(optStr string) bool {
	return (optStr != "yes" && optStr != "enabled" && optStr != "on")
}

// Listen holds the configuration for hte listeners
type Listen struct {
	Host string
	Type string

	Options map[string]string
}

// IRCOp holds o:line information
type IRCOp struct {
	User  string
	Pass  string
	Hosts []string
}

type Limits struct {
	Nick        int
	Channels    int
	ChannelName int `json:"channel_name"`
}

// Config is
type Config struct {
	Name      string
	Network   string
	LogLevel  string
	Listeners []Listen `json:"listen"`
	IrcOps    []IRCOp
	Limits    Limits `json:"limits"`
}

// New returns a new ptr of *config given a config file
// this will allow us to support multiple config files.
// though i wont say they should be needed as JSON brah
func New(configFile string) (*Config, error) {
	var bytes []byte

	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config, err := NewFromBytes(bytes)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewFromBytes(bytes []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(bytes, config); err != nil {
		return nil, err
	}

	return config, nil
}

// ListenersFor -
func (config *Config) ListenersFor(typ string) []Listen {
	listeners := []Listen{}
	for _, listener := range config.Listeners {
		if listener.Type == typ {
			listeners = append(listeners, listener)
		}
	}
	return listeners
}
