package config_test

import (
	"testing"

	"nxircd/config"
)

func TestListenersFor(t *testing.T) {
	conf, err := config.New("config.test.json")
	if err != nil {
		t.Fatal("Cannot load config.example.json for testing: ", err)
	}

	t.Run("Check IRCD Listeners", func(t *testing.T) {
		listeners := conf.ListenersFor("ircd")
		if len(listeners) != 2 {
			t.Fatalf("Expected %d to be 2", len(listeners))
		}
	})

	t.Run("Check Web Listeners", func(t *testing.T) {
		listeners := conf.ListenersFor("web")
		if len(listeners) < 1 {
			t.Fatalf("Expected %d to be greater then 0", len(listeners))
		}

		listener := listeners[0]

		auth, found := listener.Options["auth"]
		if !found {
			t.Fatalf("Expected auth options on web1 listener")
		}

		if config.IsEnabled(auth) {
			t.Fatalf("Expected auth on listner web1 to be enabled")
		}
	})
}
