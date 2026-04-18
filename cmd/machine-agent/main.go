package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/klipitkas/machine-agent/internal/discovery"
	"github.com/klipitkas/machine-agent/internal/server"
)

var version = "dev"

func main() {
	port := flag.Int("port", 7891, "port to listen on")
	showVersion := flag.Bool("version", false, "print version and exit")
	noMDNS := flag.Bool("no-mdns", false, "disable mDNS service advertisement")
	discover := flag.Bool("discover", false, "discover agents on the network and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *discover {
		fmt.Println("Scanning for agents...")
		found := 0
		err := discovery.Discover(func(e *discovery.ServiceEntry) {
			found++
			fmt.Printf("  %-20s %s:%d", e.Name, e.AddrV4, e.Port)
			if len(e.InfoFields) > 0 {
				fmt.Printf("  (%s)", e.InfoFields[0])
			}
			fmt.Println()
		})
		if err != nil {
			log.Fatalf("Discovery failed: %v", err)
		}
		if found == 0 {
			fmt.Println("No agents found.")
		} else {
			fmt.Printf("\n%d agent(s) found.\n", found)
		}
		os.Exit(0)
	}

	addr := fmt.Sprintf(":%d", *port)

	if os.Getenv("TOKEN") != "" {
		log.Println("Token authentication enabled")
	} else {
		log.Println("No TOKEN set — running without authentication")
	}

	if !*noMDNS {
		shutdown, err := discovery.Advertise(*port)
		if err != nil {
			log.Printf("mDNS advertisement failed: %v", err)
		} else {
			defer shutdown()
		}
	}

	srv := server.New(addr)

	go func() {
		log.Printf("machine-agent %s listening on %s", version, addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Stopped")
}
