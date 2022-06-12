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
	log.SetPrefix("BOOTSTRAP: ")
	log.SetFlags(0)

	var conf chaos.BootstrapConfig
	flag.StringVar(&conf.Addr, "addr", "127.0.0.1:9989", "Listen address")
	flag.Parse()

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	if err := chaos.RunBootstrap(ctx, conf); err != nil {
		log.Fatal(err)
	}
}
