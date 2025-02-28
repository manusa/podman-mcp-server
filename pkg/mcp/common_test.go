package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"net/http/httptest"
	"testing"
)

type mcpContext struct {
	ctx           context.Context
	cancel        context.CancelFunc
	mcpServer     *Server
	mcpHttpServer *httptest.Server
	mcpClient     *client.SSEMCPClient
}

func (c *mcpContext) beforeEach(t *testing.T) {
	var err error
	c.ctx, c.cancel = context.WithCancel(context.Background())
	if c.mcpServer, err = NewSever(); err != nil {
		t.Fatal(err)
		return
	}
	c.mcpHttpServer = server.NewTestServer(c.mcpServer.server)
	if c.mcpClient, err = client.NewSSEMCPClient(c.mcpHttpServer.URL + "/sse"); err != nil {
		t.Fatal(err)
		return
	}
	if err = c.mcpClient.Start(c.ctx); err != nil {
		t.Fatal(err)
		return
	}
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{Name: "test", Version: "1.33.7"}
	_, err = c.mcpClient.Initialize(c.ctx, initRequest)
	if err != nil {
		t.Fatal(err)
		return
	}
}

func (c *mcpContext) afterEach() {
	c.cancel()
	_ = c.mcpClient.Close()
	c.mcpHttpServer.Close()
}

func testCase(t *testing.T, test func(c *mcpContext)) {
	mcpCtx := &mcpContext{}
	mcpCtx.beforeEach(t)
	defer mcpCtx.afterEach()
	test(mcpCtx)
}
