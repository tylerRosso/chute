// SPDX-License-Identifier: ISC

package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestGracefulShutdown_Signals(t *testing.T) {
	signals := []syscall.Signal{syscall.SIGINT, syscall.SIGTERM}

	for _, sig := range signals {
		t.Run(sig.String(), func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("failed to listen: %v", err)
			}
			addr := listener.Addr().String()
			listener.Close()

			mux := http.NewServeMux()
			mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("pong"))
			})

			server := &http.Server{
				Addr:    addr,
				Handler: mux,
			}

			shutdown := make(chan os.Signal, 2)
			errCh := make(chan error, 1)

			go func() {
				errCh <- runServer(server, shutdown, 2*time.Second)
			}()

			// Wait for server to start accepting connections
			var resp *http.Response
			for range 50 {
				resp, err = http.Get("http://" + addr + "/ping")
				if err == nil {
					resp.Body.Close()
					break
				}
				time.Sleep(20 * time.Millisecond)
			}
			if err != nil {
				t.Fatalf("server failed to start: %v", err)
			}

			// Send signal
			shutdown <- sig

			select {
			case runErr := <-errCh:
				if runErr != nil {
					t.Errorf("expected clean shutdown, got error: %v", runErr)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("timed out waiting for server to shutdown gracefully")
			}

			// Verify server is no longer accepting new requests
			_, err = http.Get("http://" + addr + "/ping")
			if err == nil {
				t.Error("expected server to be stopped and reject connections")
			}
		})
	}
}

func TestGracefulShutdown_InFlightRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	reqStarted := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(reqStarted)
		time.Sleep(300 * time.Millisecond)
		w.Write([]byte("completed"))
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	shutdown := make(chan os.Signal, 2)
	errCh := make(chan error, 1)

	go func() {
		errCh <- runServer(server, shutdown, 2*time.Second)
	}()

	// Wait for server readiness
	for range 50 {
		conn, dialErr := net.Dial("tcp", addr)
		if dialErr == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	var clientResp string
	var clientErr error
	var wg sync.WaitGroup
	wg.Go(func() {
		resp, err := http.Get("http://" + addr + "/slow")
		if err != nil {
			clientErr = err
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		clientResp = string(body)
	})

	<-reqStarted
	// Request is now in-flight; trigger shutdown
	shutdown <- syscall.SIGTERM

	// Wait for client to finish
	wg.Wait()
	if clientErr != nil {
		t.Fatalf("in-flight request failed: %v", clientErr)
	}
	if clientResp != "completed" {
		t.Errorf("expected 'completed', got %q", clientResp)
	}

	select {
	case runErr := <-errCh:
		if runErr != nil {
			t.Errorf("expected clean shutdown, got error: %v", runErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for server to shutdown gracefully")
	}
}

func TestSetupRouter(t *testing.T) {
	tempDir := t.TempDir()
	mux := setupRouter(tempDir)
	if mux == nil {
		t.Fatal("expected non-nil mux from setupRouter")
	}
}
