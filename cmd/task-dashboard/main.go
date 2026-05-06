package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/config"
	"github.com/panda/agent-task-center/internal/handler"
	"github.com/panda/agent-task-center/internal/scanner"
	"github.com/panda/agent-task-center/web"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "Path to config YAML file")
	port := flag.Int("port", 0, "HTTP server port (overrides config file, default 8080)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", *configPath, err)
	}

	if *port != 0 {
		cfg.Server.Port = *port
	}

	runServer(cfg)
}

// runServer sets up routes and starts the HTTP server with graceful shutdown.
func runServer(cfg *config.Config) {
	s := scanner.NewScanner(cfg)
	r := setupRouter(s, cfg)

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", addr, err)
		}
		log.Printf("Task Dashboard listening on http://%s", addr)
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received %s, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

// setupRouter creates the Gin engine with all routes, middleware, and static asset serving.
func setupRouter(s *scanner.Scanner, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Serve static assets from embedded filesystem (static/ subdirectory only)
	staticSub, err := fs.Sub(web.Assets, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-filesystem: %v", err)
	}
	r.StaticFS("/static", http.FS(staticSub))

	// Register page and API routes
	handler.RegisterPages(r, s)
	handler.RegisterAPI(r, s)

	return r
}

// defaultConfigPath returns the default config path at ~/.task-dashboard.yaml.
func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".task-dashboard.yaml"
	}
	return filepath.Join(home, ".task-dashboard.yaml")
}
