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
	log.SetPrefix("NODE: ")
	log.SetFlags(0)
	var conf chaos.NodeConfig
	flag.StringVar(&conf.Addr, "addr", "127.0.0.1:0", "Listen address")
	flag.StringVar(&conf.BootstrapAddr, "bootstrap-addr", "127.0.0.1:9989", "Bootstrap address")
	flag.Parse()
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	if err := chaos.RunNode(ctx, conf); err != nil {
		log.Fatal(err)
	}
}
