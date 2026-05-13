package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bjdgyc/anylink/base"
	"github.com/bjdgyc/anylink/dbdata"
	"github.com/bjdgyc/anylink/handler"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("conf", "conf/server.toml", "config file path")
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("anylink version %s, commit %s, built at %s by %s\n",
			version, commit, date, builtBy)
		os.Exit(0)
	}

	// Initialize base configuration
	if err := base.InitConfig(*configFile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	base.InitLog()
	base.Logger.Infof("Starting anylink server %s (commit: %s)", version, commit)

	// Initialize database
	if err := dbdata.InitDB(); err != nil {
		base.Logger.Fatalf("failed to initialize database: %v", err)
	}
	defer dbdata.CloseDB()

	// Initialize and start the VPN handler
	if err := handler.Start(); err != nil {
		base.Logger.Fatalf("failed to start handler: %v", err)
	}

	// Wait for termination signal
	// Also handle SIGHUP so the process can be cleanly stopped by some init systems
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	sig := <-quit

	base.Logger.Infof("Received signal %s, shutting down...", sig)

	// Graceful shutdown
	handler.Stop()
	base.Logger.Info("anylink server stopped")
}
