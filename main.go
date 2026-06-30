package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	configFile := flag.String("config", envOr("CONFIG_FILE", "config.yml"), "config file path")
	flag.Parse()

	cfg, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	cfgStore.startCleaner()

	log.Printf("wireguard-manager listening on %s  interface=%s", cfg.ListenAddr, cfg.Interface)
	if err := http.ListenAndServe(cfg.ListenAddr, newServer(cfg)); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
