package main

import "github.com/fatih/color"

import (
  "nxircd/config"
  "nxircd/ircd"
)

import "fmt"

var red = color.New(color.FgRed).SprintFunc()

var instance *ircd.Server

func main() {
  config, err := config.New("./config.json")
  if err != nil {
    fmt.Printf("Error with configuration: %s\n", err)
    return
  }

  banner(config)

  instance = ircd.NewServer(config)
  instance.Run()

}

func banner(config *config.Config) {
  fmt.Printf(",--------------------------------------\n")
  fmt.Printf("|               %s                \n", red("NX-IRCd"))
  fmt.Printf("|--------------------------------------\n")
  fmt.Printf("| Authors: %s\n", red("Twitch"))
  fmt.Printf("| Config\n")
  fmt.Printf("|  Name: %s\n", red(config.Network))
  fmt.Printf("|  Server: %s\n", red(config.Name))
  fmt.Printf("`---------------------------------------\n")
}
