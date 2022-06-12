package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/snerror/chaos"
)

func main() {
	log.SetPrefix("CLI: ")
	log.SetFlags(0)
	var conf chaos.CliConfig
	flag.StringVar(&conf.BootstrapPath, "bootstrap-path", "bootstrap", "Bootstrap path")
	flag.StringVar(&conf.BootstrapAddr, "bootstrap-addr", "127.0.0.1:9989", "Bootstrap address")
	flag.StringVar(&conf.NodePath, "node-path", "node", "Node path")
	flag.Parse()
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	if err := chaos.RunCli(ctx, conf); err != nil {
		log.Fatal(err)
	}
}
