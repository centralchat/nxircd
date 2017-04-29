package main

import "nxircd/ircd"
import "fmt"

var instance *ircd.Server

func main() {
  config, err := ircd.NewConfig("./config.json")
  if err != nil {
    fmt.Printf("Error with configuration: %s\n", err)
    return
  }

  banner(config)

  instance = ircd.NewServer(config)
  instance.Run()

}

func banner(config *ircd.Config) {
  fmt.Printf("------------------------------------\n")
  fmt.Printf("NX-IRCd\n")
  fmt.Printf("Author: Twitch\n")
  fmt.Printf("------------------------------------\n")
}
