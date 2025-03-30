package main

import (
	"github.com/aniqp/formal_assessment/pkg/config"
	"github.com/aniqp/formal_assessment/pkg/proxy"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	sshServer, err := proxy.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(sshServer.SSHServer.ListenAndServe())
}
