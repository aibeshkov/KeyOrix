package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/secretlyhq/secretly/internal/config"
	"github.com/secretlyhq/secretly/internal/core"
	"github.com/secretlyhq/secretly/internal/i18n"
	"github.com/secretlyhq/secretly/internal/storage/local"
	"github.com/secretlyhq/secretly/internal/storage/models"
	"github.com/secretlyhq/secretly/server/grpc"
	httpServer "github.com/secretlyhq/secretly/server/http"
	"golang.org/x/crypto/acme/autocert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize i18n system
	if err := i18n.Initialize(cfg); err != nil {
		log.Fatalf("Failed to initialize i18n system: %v", err)
	}
	log.Printf("i18n system initialized with language: %s, fallback: %s", cfg.Locale.Language, cfg.Locale.FallbackLanguage)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start HTTP server
	if cfg.Server.HTTP.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := startHTTPServer(ctx, cfg); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()
	}

	// Start gRPC server
	if cfg.Server.GRPC.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := startGRPCServer(ctx, cfg); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, gracefully shutting down...")

	// Cancel context to signal shutdown
	cancel()

	// Wait for all servers to shutdown
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for graceful shutdown or timeout
	select {
	case <-done:
		log.Println("All servers shut down gracefully")
	case <-time.After(30 * time.Second):
		log.Println("Shutdown timeout exceeded, forcing exit")
	}
}

func initializeCoreService(cfg *config.Config) (*core.SecretlyCore, error) {
	// Connect to database
	db, err := gorm.Open(sqlite.Open(cfg.Storage.Database.Path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.SecretNode{}, &models.SecretVersion{}, &models.User{}, &models.Role{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize storage and core service
	storage := local.NewLocalStorage(db)
	coreService := core.NewSecretlyCore(storage)

	return coreService, nil
}

func startHTTPServer(ctx context.Context, cfg *config.Config) error {
	// Initialize core service
	coreService, err := initializeCoreService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize core service: %w", err)
	}

	// Create HTTP router
	router, err := httpServer.NewRouter(cfg, coreService)
	if err != nil {
		return fmt.Errorf("failed to create HTTP router: %w", err)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.HTTP.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure TLS if enabled
	if cfg.Server.HTTP.TLS.Enabled {
		tlsConfig, err := createTLSConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		server.TLSConfig = tlsConfig
	}

	// Start server
	go func() {
		log.Printf("Starting HTTP server on :%s", cfg.Server.HTTP.Port)
		var err error
		if cfg.Server.HTTP.TLS.Enabled {
			if cfg.Server.HTTP.TLS.AutoCert {
				// Use autocert for Let's Encrypt
				m := &autocert.Manager{
					Cache:      autocert.DirCache("certs"),
					Prompt:     autocert.AcceptTOS,
					HostPolicy: autocert.HostWhitelist(cfg.Server.HTTP.TLS.Domains...),
				}
				server.TLSConfig = m.TLSConfig()
				err = server.ListenAndServeTLS("", "")
			} else {
				// Use provided certificates
				err = server.ListenAndServeTLS(cfg.Server.HTTP.TLS.CertFile, cfg.Server.HTTP.TLS.KeyFile)
			}
		} else {
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	return server.Shutdown(shutdownCtx)
}

func startGRPCServer(ctx context.Context, cfg *config.Config) error {
	// Create gRPC server
	grpcServer, err := grpc.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	// Start server
	go func() {
		log.Printf("Starting gRPC server on :%s", cfg.Server.GRPC.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	return nil
}

func createTLSConfig(cfg *config.Config) (*tls.Config, error) {
	if cfg.Server.HTTP.TLS.AutoCert {
		// Autocert will handle TLS config
		return nil, nil
	}

	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(cfg.Server.HTTP.TLS.CertFile, cfg.Server.HTTP.TLS.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}
