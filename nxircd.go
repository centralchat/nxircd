package main

import (
	"fmt"

	"nxircd/config"
)

func main() {
	conf, err := config.New("config.json")
	if err != nil {
		fmt.Println("Unable to load in config: ", err)
		return
	}

	fmt.Println(conf)

}
