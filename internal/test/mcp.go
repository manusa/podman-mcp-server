package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	mcpServer "github.com/manusa/podman-mcp-server/pkg/mcp"
	"github.com/manusa/podman-mcp-server/pkg/podman"
)

// McpSuite is a test suite that uses a mock Podman API server
// instead of a fake podman binary. This allows testing with the real
// podman/docker CLI binary communicating with a mocked backend.
//
// Use this suite when you want to:
// - Test the actual podman/docker CLI behavior
// - Test complex API responses
// - Test error scenarios from the API
//
// Note: This suite requires a real podman or docker binary to be available.
// If neither is available, tests using this suite will be skipped.
//
// For multi-implementation testing, set the PodmanImpl field before running:
//
//	func TestContainerToolsWithCLI(t *testing.T) {
//	    suite.Run(t, &ContainerToolsSuite{
//	        McpSuite: test.McpSuite{PodmanImpl: "cli"},
//	    })
//	}
type McpSuite struct {
	suite.Suite
	// PodmanImpl specifies which Podman implementation to use for tests.
	// If empty, uses the default implementation (currently "cli").
	// Valid values: "cli" (default), "api" (future)
	PodmanImpl    string
	originalEnv   []string
	MockServer    *MockPodmanServer
	mcpServer     *mcpServer.Server
	mcpHttpServer *httptest.Server
	mcpClient     *client.Client
}

// SetupTest initializes the mock server, MCP server, and client before each test.
func (s *McpSuite) SetupTest() {
	// Check if real podman is available
	if !IsPodmanAvailable() {
		// On Linux, podman should always be available - fail the test
		// On other platforms (macOS, Windows), skip if podman is not available
		if runtime.GOOS == "linux" {
			s.Require().Fail("podman CLI is not available - install podman to run these tests")
		}
		s.T().Skip("podman CLI not available (expected on non-Linux platforms without podman machine)")
	}

	s.originalEnv = os.Environ()
	var err error

	// Start mock Podman API server
	s.MockServer = NewMockPodmanServer()

	// Set CONTAINER_HOST to point to mock server
	WithContainerHost(s.T(), s.MockServer.URL())

	// Create MCP server (it will use the real podman binary which talks to mock server)
	var serverOpts []mcpServer.ServerOption
	if s.PodmanImpl != "" {
		serverOpts = append(serverOpts, mcpServer.WithPodmanImpl(s.PodmanImpl))
	}
	s.mcpServer, err = mcpServer.NewServer(serverOpts...)
	s.Require().NoError(err)

	// Wrap in httptest.Server with Streamable HTTP handler
	streamableHandler := s.mcpServer.ServeStreamableHTTP()
	s.mcpHttpServer = httptest.NewServer(streamableHandler)

	// Create MCP client
	s.mcpClient, err = client.NewStreamableHttpClient(s.mcpHttpServer.URL)
	s.Require().NoError(err)

	err = s.mcpClient.Start(s.T().Context())
	s.Require().NoError(err)

	// Initialize MCP protocol
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{Name: "test", Version: "1.33.7"}
	_, err = s.mcpClient.Initialize(s.T().Context(), initRequest)
	s.Require().NoError(err)
}

// TearDownTest cleans up resources after each test.
func (s *McpSuite) TearDownTest() {
	if s.mcpClient != nil {
		_ = s.mcpClient.Close()
	}
	if s.mcpHttpServer != nil {
		s.mcpHttpServer.Close()
	}
	if s.MockServer != nil {
		s.MockServer.Close()
	}
	RestoreEnv(s.originalEnv)
}

// CallTool calls an MCP tool by name with the given arguments.
func (s *McpSuite) CallTool(name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	callToolRequest := mcp.CallToolRequest{}
	callToolRequest.Params.Name = name
	callToolRequest.Params.Arguments = args
	return s.mcpClient.CallTool(s.T().Context(), callToolRequest)
}

// CallToolRawResponse represents the raw JSON-RPC response from the server.
// It handles both success and error responses since JSON-RPC uses mutually exclusive
// result/error fields. The library provides separate JSONRPCResponse (success only)
// and JSONRPCError (error only) types, but we need a single struct for both cases.
type CallToolRawResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  *struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result,omitempty"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// CallToolRaw sends a raw JSON-RPC request with arbitrary arguments (as raw JSON).
// This bypasses the typed client and allows testing with malformed arguments.
// It handles the Streamable HTTP protocol including session initialization.
func (s *McpSuite) CallToolRaw(name string, argumentsJSON string) (*CallToolRawResponse, error) {
	// First, initialize the session using the same protocol version as the typed client
	initBody := fmt.Sprintf(
		`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"%s","clientInfo":{"name":"test","version":"1.0"}}}`,
		mcp.LATEST_PROTOCOL_VERSION,
	)

	initReq, err := http.NewRequest("POST", s.mcpHttpServer.URL, strings.NewReader(initBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create init request: %w", err)
	}
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Accept", "application/json, text/event-stream")

	initResp, err := http.DefaultClient.Do(initReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send init request: %w", err)
	}
	defer func() { _ = initResp.Body.Close() }()
	sessionID := initResp.Header.Get("Mcp-Session-Id")

	// Now send the tool call with the session ID
	requestBody := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"%s","arguments":%s}}`, name, argumentsJSON)

	req, err := http.NewRequest("POST", s.mcpHttpServer.URL, strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse SSE response - extract JSON from "data: " lines
	bodyStr := string(body)
	var jsonData string
	for _, line := range strings.Split(bodyStr, "\n") {
		if strings.HasPrefix(line, "data: ") {
			jsonData = strings.TrimPrefix(line, "data: ")
			break
		}
	}
	if jsonData == "" {
		// Try parsing as plain JSON (non-SSE response)
		jsonData = bodyStr
	}

	var result CallToolRawResponse
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w (body: %s)", err, bodyStr)
	}
	return &result, nil
}

// ListTools returns the list of available MCP tools.
func (s *McpSuite) ListTools() (*mcp.ListToolsResult, error) {
	return s.mcpClient.ListTools(s.T().Context(), mcp.ListToolsRequest{})
}

// WithContainerList sets up the mock server to return a list of containers.
func (s *McpSuite) WithContainerList(containers []ContainerListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, containers)
	}
	s.MockServer.HandleFunc("GET", "/libpod/containers/json", "/containers/json", handler)
}

// WithContainerInspect sets up the mock server to return container inspect data.
func (s *McpSuite) WithContainerInspect(container ContainerInspectResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, container)
	}
	s.MockServer.HandleFunc("GET", "/libpod/containers/{id}/json", "/containers/{id}/json", handler)
}

// WithImageList sets up the mock server to return a list of images.
func (s *McpSuite) WithImageList(images []ImageListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, images)
	}
	s.MockServer.HandleFunc("GET", "/libpod/images/json", "/images/json", handler)
}

// WithNetworkList sets up the mock server to return a list of networks.
func (s *McpSuite) WithNetworkList(networks []NetworkListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, networks)
	}
	s.MockServer.HandleFunc("GET", "/libpod/networks/json", "/networks", handler)
}

// WithVolumeList sets up the mock server to return a list of volumes.
// The Libpod API returns a plain array of volumes, while Docker API wraps in an object.
func (s *McpSuite) WithVolumeList(volumes []VolumeResponse) {
	// Libpod handler returns plain array
	libpodHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, volumes)
	}
	// Docker handler returns wrapped object
	dockerHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, VolumeListResponse{Volumes: volumes})
	}
	s.MockServer.Handle("GET", "/libpod/volumes/json", libpodHandler)
	s.MockServer.Handle("GET", "/volumes", dockerHandler)
	s.MockServer.Handle("GET", "/v1.40/volumes", dockerHandler)
	s.MockServer.Handle("GET", "/v1.41/volumes", dockerHandler)
}

// WithError sets up the mock server to return an error for a specific endpoint.
func (s *McpSuite) WithError(method, libpodPath, dockerPath string, statusCode int, message string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteError(w, statusCode, message)
	}
	s.MockServer.HandleFunc(method, libpodPath, dockerPath, handler)
}

// WithContainerLogs sets up the mock server to return container logs.
// The logs are encoded in the docker multiplexed stream format for non-TTY containers.
func (s *McpSuite) WithContainerLogs(logs string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		// Docker logs use multiplexed stream format for non-TTY containers
		// Format: [STREAM_TYPE][0][0][0][SIZE_BE_32][DATA...]
		// STREAM_TYPE: 0=stdin, 1=stdout, 2=stderr
		frame := make([]byte, 8+len(logs))
		frame[0] = 1 // stdout
		frame[4] = byte(len(logs) >> 24)
		frame[5] = byte(len(logs) >> 16)
		frame[6] = byte(len(logs) >> 8)
		frame[7] = byte(len(logs))
		copy(frame[8:], logs)
		_, _ = w.Write(frame)
	}
	s.MockServer.HandleFunc("GET", "/libpod/containers/{id}/logs", "/containers/{id}/logs", handler)
}

// WithContainerCreate sets up the mock server to handle container creation.
func (s *McpSuite) WithContainerCreate(containerID string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, ContainerCreateResponse{
			ID:       containerID,
			Warnings: []string{},
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/create", "/containers/create", handler)
}

// WithContainerStart sets up the mock server to handle container start.
func (s *McpSuite) WithContainerStart() {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/{id}/start", "/containers/{id}/start", handler)
}

// WithContainerStop sets up the mock server to handle container stop.
func (s *McpSuite) WithContainerStop() {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/{id}/stop", "/containers/{id}/stop", handler)
}

// WithContainerRemove sets up the mock server to handle container removal.
func (s *McpSuite) WithContainerRemove() {
	libpodHandler := func(w http.ResponseWriter, _ *http.Request) {
		// Libpod API expects a JSON array of RmReport
		WriteJSON(w, []map[string]interface{}{
			{"Id": "abc123def456"},
		})
	}
	dockerHandler := func(w http.ResponseWriter, _ *http.Request) {
		// Docker API expects HTTP 204 No Content
		w.WriteHeader(http.StatusNoContent)
	}
	s.MockServer.Handle("DELETE", "/libpod/containers/{id}", libpodHandler)
	s.MockServer.Handle("DELETE", "/containers/{id}", dockerHandler)
	s.MockServer.Handle("DELETE", "/v1.40/containers/{id}", dockerHandler)
	s.MockServer.Handle("DELETE", "/v1.41/containers/{id}", dockerHandler)
}

// WithContainerWait sets up the mock server to handle container wait.
func (s *McpSuite) WithContainerWait(exitCode int) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, map[string]interface{}{
			"StatusCode": exitCode,
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/{id}/wait", "/containers/{id}/wait", handler)
}

// WithImagePull sets up the mock server to handle image pull.
func (s *McpSuite) WithImagePull(imageID string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Podman sends streaming JSON responses during pull
		w.Header().Set("Content-Type", "application/json")
		WriteJSON(w, ImagePullResponse{
			ID:     imageID,
			Status: "Download complete",
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/images/pull", "/images/create", handler)
}

// WithImageRemove sets up the mock server to handle image removal.
// The Libpod API uses /libpod/images/remove with image names as query params.
func (s *McpSuite) WithImageRemove(imageID string) {
	// Libpod API handler - returns LibpodImagesRemoveReport
	libpodHandler := func(w http.ResponseWriter, _ *http.Request) {
		// Response format: {"Deleted": ["sha256:..."], "Untagged": [...], "Errors": [...], "ExitCode": 0}
		WriteJSON(w, map[string]any{
			"Deleted":  []string{imageID},
			"Untagged": []string{},
			"Errors":   []string{},
			"ExitCode": 0,
		})
	}
	// Docker API handler - returns array of ImageRemoveResponse
	dockerHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, []ImageRemoveResponse{
			{Deleted: imageID},
		})
	}
	s.MockServer.Handle("DELETE", "/libpod/images/remove", libpodHandler)
	s.MockServer.Handle("DELETE", "/images/{name}", dockerHandler)
	s.MockServer.Handle("DELETE", "/v1.40/images/{name}", dockerHandler)
	s.MockServer.Handle("DELETE", "/v1.41/images/{name}", dockerHandler)
}

// WithImagePush sets up the mock server to handle image push.
func (s *McpSuite) WithImagePush() {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		// Podman sends streaming JSON responses during push
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Empty response body is acceptable for push
	}
	s.MockServer.HandleFunc("POST", "/libpod/images/{name}/push", "/images/{name}/push", handler)
}

// WithContainerRun sets up the mock server to handle container creation and start.
// Returns a container ID and sets up all required endpoints for `podman run`.
func (s *McpSuite) WithContainerRun(containerID string) {
	// Image pull (podman always tries to pull first)
	pullHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, ImagePullResponse{
			ID:     "sha256:abc123def456",
			Status: "Already exists",
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/images/pull", "/images/create", pullHandler)

	// Image inspect (podman inspects the image after pull)
	imageInspectHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, map[string]any{
			"Id":       "sha256:abc123def456",
			"RepoTags": []string{"example.com/org/image:tag"},
			"Config": map[string]any{
				"ExposedPorts": map[string]any{
					"80/tcp": map[string]any{},
				},
			},
		})
	}
	s.MockServer.Handle("GET", "/libpod/images/{name}/json", imageInspectHandler)

	// Container create
	createHandler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, map[string]string{
			"Id": containerID,
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/create", "/containers/create", createHandler)

	// Container start
	startHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	s.MockServer.Handle("POST", "/libpod/containers/{id}/start", startHandler)
	s.MockServer.Handle("POST", "/containers/{id}/start", startHandler)
}

// WithImageBuild sets up the mock server to handle image builds.
// Returns a successful build response.
func (s *McpSuite) WithImageBuild(imageID string) {
	buildHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return streaming build output in NDJSON format
		// Note: Use \\n for escaped newline in JSON, \n for line separator
		_, _ = w.Write([]byte(`{"stream":"Step 1/1 : FROM alpine\\n"}` + "\n"))
		_, _ = w.Write([]byte(`{"stream":"Successfully built ` + imageID + `\\n"}` + "\n"))
		// Write ID in format podman expects
		_, _ = w.Write([]byte(`{"stream":"` + imageID + `\\n"}` + "\n"))
	}
	s.MockServer.Handle("POST", "/libpod/build", buildHandler)
	s.MockServer.Handle("POST", "/build", buildHandler)
}

// GetCapturedRequest returns the first captured request matching the method and path pattern.
// Returns nil if no matching request is found.
func (s *McpSuite) GetCapturedRequest(method, pathPattern string) *CapturedRequest {
	return s.MockServer.GetRequest(method, pathPattern)
}

// PopLastCapturedRequest returns the last captured request matching the method and path pattern,
// and removes it from the captured requests stack. This allows multiple subtests within
// the same test function to each get their own captured request.
// Returns nil if no matching request is found.
func (s *McpSuite) PopLastCapturedRequest(method, pathPattern string) *CapturedRequest {
	return s.MockServer.PopLastRequest(method, pathPattern)
}

// AvailableImplementations returns the list of Podman implementations available for testing.
// This can be used to create parameterized tests that run the same suite with different implementations.
//
// Example usage with testify/suite:
//
//	func TestContainerSuiteWithAllImplementations(t *testing.T) {
//	    for _, impl := range test.AvailableImplementations() {
//	        suite.Run(t, &ContainerSuite{
//	            McpSuite: test.McpSuite{PodmanImpl: impl},
//	        })
//	    }
//	}
func AvailableImplementations() []string {
	// Returns all implementations registered in the registry.
	// Currently only "cli" is registered. "api" will be added in Phase 2.
	return podman.ImplementationNames()
}

// DefaultImplementation returns the default Podman implementation for testing.
// This is the implementation used when PodmanImpl is empty.
func DefaultImplementation() string {
	// Returns "cli" as the default, which matches the auto-detection behavior
	// when only CLI is registered (CLI has priority 50).
	return "cli"
}
