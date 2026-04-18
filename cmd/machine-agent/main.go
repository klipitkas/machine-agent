package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/klipitkas/machine-agent/internal/server"
)

func main() {
	port := flag.Int("port", 7891, "port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	if os.Getenv("TOKEN") != "" {
		log.Println("Token authentication enabled")
	} else {
		log.Println("No TOKEN set — running without authentication")
	}

	srv := server.New(addr)

	go func() {
		log.Printf("machine-agent listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}
