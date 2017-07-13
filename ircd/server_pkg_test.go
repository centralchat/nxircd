package ircd_test

import (
	"nxircd/config"
	"nxircd/ircd"
	"os"
	"time"

	logger "github.com/apsdehal/go-logger"
)

var exampleConfig = `
{
  "name":    "irc.nxircd.org",
  "network": "nxircd",
  
  "loglevel": "DEBUG",
  "listen": [
    { "host": "127.0.0.1:6666", "type" : "ircd" },
    { "host": "127.0.0.1:6667", "type" : "ircd" },
    { "host": "127.0.0.1:9001", "type" : "ws" },
    { 
       "host": ":8080",         
       "type":  "web",   
       "options": {
         "auth": "enabled"
       }
    }
  ],  
  "ircops": [
    {
      "user": "developer",
      "pass": "password",
      "hosts": [
        "*.example.com",
        "localhost"
      ]
    }
  ]
}
`

func NewTestServer() *ircd.Server {
	conf, _ := config.NewFromBytes([]byte(exampleConfig))
	log, _ := logger.New("test", 1, os.Stdout)

	return &ircd.Server{
		Name:    "testing",
		Network: "irc.testing.tst",

		CTime: time.Now(),
		Time:  time.Now(),

		Log:    log,
		Config: conf,

		Clients:  ircd.NewClientList(),
		Channels: ircd.NewChanList(),

		MaxClients: 0,

		Signals: make(chan os.Signal),
	}
}
