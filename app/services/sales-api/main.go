package main

import (
	"go.uber.org/automaxprocs/maxprocs"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var build = "develop"

func main() {
	if _, err := maxprocs.Set(); err != nil {
		log.Printf("maxprocs: %v\n", err)
		os.Exit(1)
	}

	g := runtime.GOMAXPROCS(0)

	log.Printf("starting service build[%s] CPU[%d]", build, g)
	defer log.Println("service ended")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Println("stopping service")
}