package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// MockPodmanServer is a mock HTTP server that simulates the Podman/Docker API.
// It supports both Libpod API (/libpod/...) and Docker-compatible API (/v1.x/... and /...).
type MockPodmanServer struct {
	server   *httptest.Server
	handlers map[string]http.HandlerFunc
	requests []CapturedRequest
	mu       sync.RWMutex
}

// CapturedRequest represents a captured HTTP request for test assertions.
type CapturedRequest struct {
	Method string
	Path   string
	Query  string
	Body   []byte
}

// NewMockPodmanServer creates a new mock Podman API server.
func NewMockPodmanServer() *MockPodmanServer {
	m := &MockPodmanServer{
		handlers: make(map[string]http.HandlerFunc),
		requests: make([]CapturedRequest, 0),
	}
	m.server = httptest.NewServer(http.HandlerFunc(m.ServeHTTP))
	m.registerBaseEndpoints()
	return m
}

// ServeHTTP handles incoming HTTP requests by routing to registered handlers.
func (m *MockPodmanServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Capture the request
	m.captureRequest(r)

	m.mu.RLock()
	defer m.mu.RUnlock()

	path := r.URL.Path

	// Try exact match first
	if handler, ok := m.handlers[r.Method+" "+path]; ok {
		handler(w, r)
		return
	}

	// Try pattern matching for paths with IDs (e.g., /containers/{id}/json)
	for pattern, handler := range m.handlers {
		if matchPath(pattern, r.Method+" "+path) {
			handler(w, r)
			return
		}
	}

	// Default 404 response
	http.NotFound(w, r)
}

// captureRequest captures an incoming request for later inspection.
func (m *MockPodmanServer) captureRequest(r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	captured := CapturedRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
	}
	// Note: We don't read the body here to avoid consuming it before handlers
	m.requests = append(m.requests, captured)
}

// Requests returns all captured requests.
func (m *MockPodmanServer) Requests() []CapturedRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]CapturedRequest{}, m.requests...)
}

// LastRequest returns the most recent captured request, or nil if none.
func (m *MockPodmanServer) LastRequest() *CapturedRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.requests) == 0 {
		return nil
	}
	req := m.requests[len(m.requests)-1]
	return &req
}

// ClearRequests clears all captured requests.
func (m *MockPodmanServer) ClearRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = make([]CapturedRequest, 0)
}

// HasRequest checks if a request with the given method and path pattern was made.
func (m *MockPodmanServer) HasRequest(method, pathPattern string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, req := range m.requests {
		if req.Method == method && matchPathSimple(pathPattern, req.Path) {
			return true
		}
	}
	return false
}

// matchPathSimple checks if a path matches a pattern (supports {placeholder} syntax).
func matchPathSimple(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, pp := range patternParts {
		if strings.HasPrefix(pp, "{") && strings.HasSuffix(pp, "}") {
			continue
		}
		if pp != pathParts[i] {
			return false
		}
	}
	return true
}

// matchPath checks if a request path matches a pattern with placeholders.
// Pattern example: "GET /libpod/containers/{id}/json"
// Path example: "GET /libpod/containers/abc123/json"
func matchPath(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, pp := range patternParts {
		if strings.HasPrefix(pp, "{") && strings.HasSuffix(pp, "}") {
			// Placeholder matches any value
			continue
		}
		if pp != pathParts[i] {
			return false
		}
	}
	return true
}

// URL returns the base URL of the mock server.
func (m *MockPodmanServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server.
func (m *MockPodmanServer) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// Handle registers a handler for a specific HTTP method and path.
// The path can include placeholders like {id} for dynamic segments.
func (m *MockPodmanServer) Handle(method, path string, handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[method+" "+path] = handler
}

// HandleFunc is a convenience method that registers a handler for both
// Libpod and Docker-compatible API paths.
func (m *MockPodmanServer) HandleFunc(method, libpodPath, dockerPath string, handler http.HandlerFunc) {
	m.Handle(method, libpodPath, handler)
	if dockerPath != "" {
		m.Handle(method, dockerPath, handler)
		// Also register with version prefix for Docker API
		m.Handle(method, "/v1.40"+dockerPath, handler)
		m.Handle(method, "/v1.41"+dockerPath, handler)
	}
}

// ResetHandlers clears all registered handlers and re-registers base endpoints.
// Also clears captured requests.
func (m *MockPodmanServer) ResetHandlers() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = make(map[string]http.HandlerFunc)
	m.requests = make([]CapturedRequest, 0)
	m.registerBaseEndpointsLocked()
}

// registerBaseEndpoints registers the base API endpoints required for podman/docker CLI.
func (m *MockPodmanServer) registerBaseEndpoints() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registerBaseEndpointsLocked()
}

// registerBaseEndpointsLocked registers base endpoints (must hold lock).
func (m *MockPodmanServer) registerBaseEndpointsLocked() {
	// Health check endpoint
	pingHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
	m.handlers["GET /_ping"] = pingHandler
	m.handlers["HEAD /_ping"] = pingHandler
	m.handlers["GET /libpod/_ping"] = pingHandler

	// Version endpoint
	versionHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, VersionResponse{
			APIVersion:    "1.40",
			Arch:          "amd64",
			Built:         1234567890,
			BuildTime:     "2024-01-01T00:00:00Z",
			Components:    []ComponentVersion{},
			Experimental:  false,
			GitCommit:     "mock",
			GoVersion:     "go1.21",
			KernelVersion: "5.15.0",
			MinAPIVersion: "1.24",
			Os:            "linux",
			Version:       "4.0.0",
		})
	}
	m.handlers["GET /version"] = versionHandler
	m.handlers["GET /libpod/version"] = versionHandler
	m.handlers["GET /v1.40/version"] = versionHandler
	m.handlers["GET /v1.41/version"] = versionHandler

	// Info endpoint
	infoHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, InfoResponse{
			Host: HostInfo{
				Arch:           "amd64",
				BuildahVersion: "1.30.0",
				Conmon: ConmonInfo{
					Package: "conmon-2.1.0",
					Path:    "/usr/bin/conmon",
					Version: "conmon version 2.1.0",
				},
				Distribution: DistributionInfo{
					Distribution: "fedora",
					Version:      "39",
				},
				Hostname:    "mock-host",
				Kernel:      "5.15.0",
				MemFree:     1024 * 1024 * 1024,
				MemTotal:    8 * 1024 * 1024 * 1024,
				OS:          "linux",
				Rootless:    true,
				Uptime:      "1h 0m 0s",
				CPUs:        4,
				EventLogger: "journald",
				SecurityInfo: SecurityInfo{
					AppArmorEnabled:    false,
					SELinuxEnabled:     false,
					SeccompEnabled:     true,
					SeccompProfilePath: "/usr/share/containers/seccomp.json",
					Rootless:           true,
				},
			},
			Store: StoreInfo{
				GraphDriverName: "overlay",
				GraphRoot:       "/home/user/.local/share/containers/storage",
				RunRoot:         "/run/user/1000/containers",
				ImageStore: ImageStoreInfo{
					Number: 10,
				},
				ContainerStore: ContainerStoreInfo{
					Number:  5,
					Running: 2,
					Paused:  0,
					Stopped: 3,
				},
			},
			Version: VersionInfo{
				APIVersion: "1.40",
				Built:      1234567890,
				BuiltTime:  "2024-01-01T00:00:00Z",
				GitCommit:  "mock",
				GoVersion:  "go1.21",
				OsArch:     "linux/amd64",
				Version:    "4.0.0",
			},
		})
	}
	m.handlers["GET /info"] = infoHandler
	m.handlers["GET /libpod/info"] = infoHandler
	m.handlers["GET /v1.40/info"] = infoHandler
	m.handlers["GET /v1.41/info"] = infoHandler
}

// WriteJSON writes a JSON response to the http.ResponseWriter.
func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// WriteError writes an error response in the format expected by podman/docker CLI.
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Cause:    message,
		Message:  message,
		Response: statusCode,
	})
}
