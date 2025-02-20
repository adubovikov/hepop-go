package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sipcapture/hepop-go/internal/api"
	"github.com/sipcapture/hepop-go/internal/config"
	"github.com/sipcapture/hepop-go/internal/writer"
)

func main() {
	// load configuration
	configFile := "config/config.yaml"
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}

	// initialize writer
	hepWriter, err := initializeWriter(cfg)
	if err != nil {
		log.Fatalf("error initializing writer: %v", err)
	}
	defer func() {
		if err := hepWriter.Close(); err != nil {
			log.Printf("error closing writer: %v", err)
		}
	}()
	// start API
	apiServer := api.NewAPI(&api.Config{
		Host:        cfg.API.Host,
		Port:        cfg.API.Port,
		EnablePprof: cfg.API.EnablePprof,
		AuthToken:   cfg.API.AuthToken,
	}, hepWriter)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("error starting API: %v", err)
		}
	}()

	// wait for API shutdown
	defer func() {
		if err := apiServer.Stop(); err != nil {
			log.Printf("error stopping API: %v", err)
		}
	}()

	// wait for shutdown
	waitForShutdown(apiServer)
}

// initializeWriter initializes the writer based on the configuration
func initializeWriter(cfg *config.Config) (writer.Writer, error) {
	switch cfg.Writers.Type {
	case "clickhouse":
		return writer.NewClickHouseWriter(writer.ClickHouseConfig{
			Host:     cfg.Writers.ClickHouse.Host,
			Port:     cfg.Writers.ClickHouse.Port,
			Database: cfg.Writers.ClickHouse.Database,
			Table:    cfg.Writers.ClickHouse.Table,
			Username: cfg.Writers.ClickHouse.Username,
			Password: cfg.Writers.ClickHouse.Password,
		})
	case "elastic":
		return writer.NewElasticWriter(writer.ElasticConfig{})
	// Add other writer types if necessary
	default:
		return nil, fmt.Errorf("unknown writer type: %s", cfg.Writers.Type)
	}
}

// waitForShutdown handles shutdown signals and gracefully stops the server
func waitForShutdown(apiServer *api.API) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("received signal %s, shutting down...\n", sig)

	log.Println("server stopped.")
}
