package main

import (
	"fmt"
	"log"

	"github.com/alandtsang/gocode/config"
)

func main() {
	// load config
	config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// print config
	fmt.Printf("config: %+v\n", config)
}
