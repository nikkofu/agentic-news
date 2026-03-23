package main

import (
	"net/http"
	"testing"
	"time"
)

func TestNewServerConfiguresTimeouts(t *testing.T) {
	server := newServer(":8081", http.NewServeMux())

	if server.Addr != ":8081" {
		t.Fatalf("unexpected addr: %q", server.Addr)
	}
	if server.Handler == nil {
		t.Fatal("expected handler")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("unexpected ReadHeaderTimeout: %v", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("unexpected ReadTimeout: %v", server.ReadTimeout)
	}
	if server.WriteTimeout != 15*time.Second {
		t.Fatalf("unexpected WriteTimeout: %v", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected IdleTimeout: %v", server.IdleTimeout)
	}
}
