// Package main is the entry point for the relayctl reverse proxy manager.
// It wires together configuration loading, proxy serving, admin API,
// health checking, metrics, and hot-reload via file watching.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/relayctl/internal/admin"
	"github.com/yourusername/relayctl/internal/circuitbreaker"
	"github.com/yourusername/relayctl/internal/config"
	"github.com/yourusername/relayctl/internal/healthcheck"
	"github.com/yourusername/relayctl/internal/metrics"
	"github.com/yourusername/relayctl/internal/proxy"
	"github.com/yourusername/relayctl/internal/ratelimit"
	"github.com/yourusername/relayctl/internal/retry"
	"github.com/yourusername/relayctl/internal/watcher"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Load initial configuration.
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("relayctl: failed to load config: %v", err)
	}

	// Build shared components.
	m := metrics.New()
	cbRegistry := circuitbreaker.NewRegistry(circuitbreaker.RegistryConfig{
		MaxFailures:     5,
		OpenTimeout:     30 * time.Second,
		HalfOpenMaxReqs: 2,
	})
	hc := healthcheck.New(healthcheck.Config{
		Interval: 15 * time.Second,
		Timeout:  5 * time.Second,
	})
	rl := ratelimit.New(ratelimit.Config{
		Requests: 100,
		Window:   time.Second,
	})
	retryPolicy := retry.DefaultPolicy()

	// Register backends with health checker.
	for _, route := range cfg.Routes {
		hc.Register(route.Backend)
	}
	hc.Start()
	defer hc.Stop()

	// Build proxy handler.
	p, err := proxy.New(cfg, m, cbRegistry, hc, rl, retryPolicy)
	if err != nil {
		log.Fatalf("relayctl: failed to create proxy: %v", err)
	}

	// Start proxy server.
	proxyServer := proxy.NewServer(cfg.Addr, p)
	go func() {
		log.Printf("relayctl: proxy listening on %s", cfg.Addr)
		if err := proxyServer.Start(); err != nil {
			log.Printf("relayctl: proxy server stopped: %v", err)
		}
	}()

	// Set up hot-reload via config file watcher.
	reloader := watcher.NewReloader(*cfgPath, func(newCfg *config.Config) error {
		if err := p.Reload(newCfg); err != nil {
			return err
		}
		if err := proxyServer.UpdateAddr(newCfg.Addr); err != nil {
			log.Printf("relayctl: addr update skipped: %v", err)
		}
		return nil
	})

	fw, err := watcher.New(*cfgPath, reloader.Reload)
	if err != nil {
		log.Fatalf("relayctl: failed to start file watcher: %v", err)
	}
	defer fw.Stop()

	// Build and start admin API server.
	adminHandler := admin.NewHandler(cfg, reloader)
	adminHandler.Register("/metrics", admin.NewMetricsHandler(m))
	adminHandler.Register("/ratelimit", admin.NewRateLimitHandler(rl))
	adminHandler.Register("/healthcheck", admin.NewHealthCheckHandler(hc))
	adminHandler.Register("/retry", admin.NewRetryHandler(retryPolicy))
	adminHandler.Register("/circuitbreaker", admin.NewCircuitBreakerRegistryHandler(cbRegistry))
	adminHandler.Register("/routes", admin.NewRoutesHandler(cfg))

	adminServer := admin.NewServer(cfg.AdminAddr, adminHandler)
	go func() {
		log.Printf("relayctl: admin API listening on %s", cfg.AdminAddr)
		if err := adminServer.Start(); err != nil {
			log.Printf("relayctl: admin server stopped: %v", err)
		}
	}()

	// Wait for termination signal.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("relayctl: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := proxyServer.Shutdown(ctx); err != nil {
		log.Printf("relayctl: proxy shutdown error: %v", err)
	}
	if err := adminServer.Shutdown(ctx); err != nil {
		log.Printf("relayctl: admin shutdown error: %v", err)
	}

	log.Println("relayctl: stopped")
}
