package podman

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ErrNoSocketFound is returned when no Podman socket can be detected.
var ErrNoSocketFound = errors.New("no podman socket found")

// DetectSocket returns the first available Podman socket path.
// It checks in the following order:
// 1. CONTAINER_HOST environment variable
// 2. Rootful default: /run/podman/podman.sock
// 3. Rootless default: $XDG_RUNTIME_DIR/podman/podman.sock
// 4. Rootless fallback: /run/user/<UID>/podman/podman.sock
//
// Returns the socket URI (e.g., "unix:///run/podman/podman.sock") or an error.
func DetectSocket() (string, error) {
	// Check CONTAINER_HOST environment variable first
	if containerHost := os.Getenv("CONTAINER_HOST"); containerHost != "" {
		// Validate it's a valid URI
		if _, err := url.Parse(containerHost); err != nil {
			return "", fmt.Errorf("invalid CONTAINER_HOST: %w", err)
		}
		return containerHost, nil
	}

	// Build list of socket paths to check
	socketPaths := []string{
		"/run/podman/podman.sock", // Rootful default
	}

	// Add rootless paths
	if xdgRuntimeDir := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntimeDir != "" {
		socketPaths = append(socketPaths, filepath.Join(xdgRuntimeDir, "podman", "podman.sock"))
	}

	// Add rootless fallback using UID
	uid := os.Getuid()
	socketPaths = append(socketPaths, "/run/user/"+strconv.Itoa(uid)+"/podman/podman.sock")

	// Check each path
	for _, path := range socketPaths {
		if _, err := os.Stat(path); err == nil {
			return "unix://" + path, nil
		}
	}

	return "", ErrNoSocketFound
}

// PingSocket verifies that the socket at the given URI responds to API requests.
// It performs a simple connection test followed by a HEAD request to /_ping.
func PingSocket(socketURI string) error {
	u, err := url.Parse(socketURI)
	if err != nil {
		return fmt.Errorf("invalid socket URI: %w", err)
	}

	var network, address string
	switch u.Scheme {
	case "unix":
		network = "unix"
		address = u.Path
	case "tcp":
		network = "tcp"
		address = u.Host
	default:
		return fmt.Errorf("unsupported socket scheme: %s", u.Scheme)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to connect to the socket
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return fmt.Errorf("failed to connect to socket: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Send a simple HTTP HEAD request to /_ping
	// This validates the socket is actually a Podman API endpoint
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	request := "HEAD /_ping HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
	if _, err := conn.Write([]byte(request)); err != nil {
		return fmt.Errorf("failed to send ping request: %w", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read ping response: %w", err)
	}

	response := string(buf[:n])
	if !strings.Contains(response, "200 OK") {
		return fmt.Errorf("unexpected ping response: %s", response)
	}

	return nil
}
