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
	podmanBinaryDir string
	mcpServer       *mcpServer.Server
	mcpHttpServer   *httptest.Server
	mcpClient       *client.Client
}

// SetupTest initializes the MCP server and client before each test.
func (s *McpSuite) SetupTest() {
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
	MockServer    *MockPodmanServer
	mcpServer     *mcpServer.Server
	mcpHttpServer *httptest.Server
	mcpClient     *client.Client
	cleanupEnv    func()
}

// SetupTest initializes the mock server, MCP server, and client before each test.
func (s *MockServerMcpSuite) SetupTest() {
	// Check if real podman or docker is available
	s.Require().True(IsPodmanAvailable() || IsDockerAvailable(),
		"neither podman nor docker CLI is available - install one to run these tests")

	var err error

	// Start mock Podman API server
	s.MockServer = NewMockPodmanServer()

	// Set CONTAINER_HOST/DOCKER_HOST to point to mock server
	s.cleanupEnv = WithContainerHost(s.T(), s.MockServer.URL())

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
	if s.cleanupEnv != nil {
		s.cleanupEnv()
	}
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
