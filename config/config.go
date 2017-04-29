package config

import (
  "encoding/json"
  "io/ioutil"
)

// ListenConfig holds the configuration for hte listeners
type ListenConfig struct {
  Host string
  Type string
}

// IrcOpConfig holds o:line information
type IrcOpConfig struct {
  User  string
  Pass  string
  Hosts []string
}

// Config is
type Config struct {
  Name     string
  Network  string
  LogLevel string
  Listen   []ListenConfig
  IrcOps   []IrcOpConfig
}

// New is
func New(configFile string) (config *Config, err error) {
  var bytes []byte

  bytes, err = ioutil.ReadFile(configFile)
  if err != nil {
    return
  }

  if err = json.Unmarshal(bytes, &config); err != nil {
    return
  }

  return
}

func (config *Config) ListenersFor(typ string) (listeners []ListenConfig) {
  i := 0

  listeners = make([]ListenConfig, len(config.Listen))
  for _, listener := range config.Listen {
    if listener.Type == typ {
      listeners[i] = listener
      i++
    }
  }
  return
}
