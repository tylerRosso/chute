// SPDX-License-Identifier: ISC

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mdp/qrterminal/v3"
)

func runServer(server *http.Server, shutdown <-chan os.Signal, timeout time.Duration) error {
	serverErrors := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case sig := <-shutdown:
		log.Printf("Received %v signal, shutting down gracefully...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		go func() {
			select {
			case <-shutdown:
				log.Println("Received second signal, forcing immediate shutdown.")
				server.Close()
			case <-ctx.Done():
			}
		}()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during graceful shutdown: %v", err)
			return err
		}
		log.Println("Server stopped cleanly.")
		return nil
	}
}

func main() {
	portFlag := flag.Int("port", 8080, "Port to run the HTTP server on")
	dirFlag := flag.String("dir", "./uploads", "Directory where uploaded files will be stored")
	noQRFlag := flag.Bool("no-qr", false, "Disable printing the QR code in the terminal")
	flag.Parse()

	uploadDir, err := filepath.Abs(*dirFlag)
	if err != nil {
		log.Fatalf("Invalid directory path: %v", err)
	}
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	mux := setupRouter(uploadDir)

	portStr := fmt.Sprintf(":%d", *portFlag)
	ip := getLocalIP()
	localURL := fmt.Sprintf("http://localhost:%d", *portFlag)
	networkURL := fmt.Sprintf("http://%s:%d", ip, *portFlag)

	fmt.Println("====================================================")
	fmt.Printf(" chute running!\n")
	fmt.Printf(" Storage Path: %s\n", uploadDir)
	fmt.Printf(" Desktop URL:  %s\n", localURL)
	fmt.Printf(" Mobile URL:   %s\n", networkURL)
	fmt.Println("====================================================")

	if !*noQRFlag {
		fmt.Println("\nScan with your phone camera:")
		config := qrterminal.Config{
			Level:      qrterminal.L,
			Writer:     os.Stdout,
			HalfBlocks: true,
		}
		qrterminal.GenerateWithConfig(networkURL, config)
		fmt.Println()
	}

	server := &http.Server{
		Addr:    portStr,
		Handler: mux,
	}

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	if err := runServer(server, shutdown, 10*time.Second); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
