package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	mcpServer "github.com/manusa/podman-mcp-server/pkg/mcp"
)

// McpSuite is a base test suite for MCP server tests.
// Embed this suite in your test suites to get MCP client helpers.
type McpSuite struct {
	suite.Suite
	originalEnv     []string
	podmanBinaryDir string
	mcpServer       *mcpServer.Server
	mcpHttpServer   *httptest.Server
	mcpClient       *client.Client
}

// SetupTest initializes the MCP server and client before each test.
func (s *McpSuite) SetupTest() {
	s.originalEnv = os.Environ()
	var err error
	s.podmanBinaryDir = WithPodmanBinary(s.T())
	s.mcpServer, err = mcpServer.NewServer()
	s.Require().NoError(err)
	// Use the go-sdk's Streamable HTTP handler wrapped in httptest.Server
	streamableHandler := s.mcpServer.ServeStreamableHTTP()
	s.mcpHttpServer = httptest.NewServer(streamableHandler)
	s.mcpClient, err = client.NewStreamableHttpClient(s.mcpHttpServer.URL)
	s.Require().NoError(err)
	err = s.mcpClient.Start(s.T().Context())
	s.Require().NoError(err)
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{Name: "test", Version: "1.33.7"}
	_, err = s.mcpClient.Initialize(s.T().Context(), initRequest)
	s.Require().NoError(err)
}

// TearDownTest cleans up the MCP server and client after each test.
func (s *McpSuite) TearDownTest() {
	_ = s.mcpClient.Close()
	s.mcpHttpServer.Close()
	RestoreEnv(s.originalEnv)
}

// CallTool calls an MCP tool by name with the given arguments.
func (s *McpSuite) CallTool(name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	callToolRequest := mcp.CallToolRequest{}
	callToolRequest.Params.Name = name
	callToolRequest.Params.Arguments = args
	return s.mcpClient.CallTool(s.T().Context(), callToolRequest)
}

// WithPodmanOutput sets up mock output for the fake podman binary.
func (s *McpSuite) WithPodmanOutput(outputLines ...string) {
	if len(outputLines) > 0 {
		f, _ := os.Create(path.Join(s.podmanBinaryDir, "output.txt"))
		defer func() { _ = f.Close() }()
		for _, line := range outputLines {
			_, _ = f.WriteString(line + "\n")
		}
	}
}

// ListTools returns the list of available MCP tools.
func (s *McpSuite) ListTools() (*mcp.ListToolsResult, error) {
	return s.mcpClient.ListTools(s.T().Context(), mcp.ListToolsRequest{})
}

// MockServerMcpSuite is a test suite that uses a mock Podman API server
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
type MockServerMcpSuite struct {
	suite.Suite
	originalEnv   []string
	MockServer    *MockPodmanServer
	mcpServer     *mcpServer.Server
	mcpHttpServer *httptest.Server
	mcpClient     *client.Client
}

// SetupTest initializes the mock server, MCP server, and client before each test.
func (s *MockServerMcpSuite) SetupTest() {
	// Check if real podman is available
	s.Require().True(IsPodmanAvailable(),
		"podman CLI is not available - install podman to run these tests")

	s.originalEnv = os.Environ()
	var err error

	// Start mock Podman API server
	s.MockServer = NewMockPodmanServer()

	// Set CONTAINER_HOST to point to mock server
	WithContainerHost(s.T(), s.MockServer.URL())

	// Create MCP server (it will use the real podman binary which talks to mock server)
	s.mcpServer, err = mcpServer.NewServer()
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
func (s *MockServerMcpSuite) TearDownTest() {
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
func (s *MockServerMcpSuite) CallTool(name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	callToolRequest := mcp.CallToolRequest{}
	callToolRequest.Params.Name = name
	callToolRequest.Params.Arguments = args
	return s.mcpClient.CallTool(s.T().Context(), callToolRequest)
}

// ListTools returns the list of available MCP tools.
func (s *MockServerMcpSuite) ListTools() (*mcp.ListToolsResult, error) {
	return s.mcpClient.ListTools(s.T().Context(), mcp.ListToolsRequest{})
}

// WithContainerList sets up the mock server to return a list of containers.
func (s *MockServerMcpSuite) WithContainerList(containers []ContainerListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, containers)
	}
	s.MockServer.HandleFunc("GET", "/libpod/containers/json", "/containers/json", handler)
}

// WithContainerInspect sets up the mock server to return container inspect data.
func (s *MockServerMcpSuite) WithContainerInspect(container ContainerInspectResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, container)
	}
	s.MockServer.HandleFunc("GET", "/libpod/containers/{id}/json", "/containers/{id}/json", handler)
}

// WithImageList sets up the mock server to return a list of images.
func (s *MockServerMcpSuite) WithImageList(images []ImageListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, images)
	}
	s.MockServer.HandleFunc("GET", "/libpod/images/json", "/images/json", handler)
}

// WithNetworkList sets up the mock server to return a list of networks.
func (s *MockServerMcpSuite) WithNetworkList(networks []NetworkListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, networks)
	}
	s.MockServer.HandleFunc("GET", "/libpod/networks/json", "/networks", handler)
}

// WithVolumeList sets up the mock server to return a list of volumes.
func (s *MockServerMcpSuite) WithVolumeList(volumes VolumeListResponse) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, volumes)
	}
	s.MockServer.HandleFunc("GET", "/libpod/volumes/json", "/volumes", handler)
}

// WithError sets up the mock server to return an error for a specific endpoint.
func (s *MockServerMcpSuite) WithError(method, libpodPath, dockerPath string, statusCode int, message string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteError(w, statusCode, message)
	}
	s.MockServer.HandleFunc(method, libpodPath, dockerPath, handler)
}

// WithContainerLogs sets up the mock server to return container logs.
// The logs are encoded in the docker multiplexed stream format for non-TTY containers.
func (s *MockServerMcpSuite) WithContainerLogs(logs string) {
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
func (s *MockServerMcpSuite) WithContainerCreate(containerID string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, ContainerCreateResponse{
			ID:       containerID,
			Warnings: []string{},
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/create", "/containers/create", handler)
}

// WithContainerStart sets up the mock server to handle container start.
func (s *MockServerMcpSuite) WithContainerStart() {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/{id}/start", "/containers/{id}/start", handler)
}

// WithContainerStop sets up the mock server to handle container stop.
func (s *MockServerMcpSuite) WithContainerStop() {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/{id}/stop", "/containers/{id}/stop", handler)
}

// WithContainerRemove sets up the mock server to handle container removal.
func (s *MockServerMcpSuite) WithContainerRemove() {
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
func (s *MockServerMcpSuite) WithContainerWait(exitCode int) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, map[string]interface{}{
			"StatusCode": exitCode,
		})
	}
	s.MockServer.HandleFunc("POST", "/libpod/containers/{id}/wait", "/containers/{id}/wait", handler)
}

// WithImagePull sets up the mock server to handle image pull.
func (s *MockServerMcpSuite) WithImagePull(imageID string) {
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
func (s *MockServerMcpSuite) WithImageRemove(imageID string) {
	handler := func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, []ImageRemoveResponse{
			{Deleted: imageID},
		})
	}
	s.MockServer.HandleFunc("DELETE", "/libpod/images/{name}", "/images/{name}", handler)
}
