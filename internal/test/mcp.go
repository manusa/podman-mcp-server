package test

import (
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
