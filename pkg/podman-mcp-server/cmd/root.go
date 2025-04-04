package cmd

import (
	"errors"
	"fmt"
	"github.com/manusa/podman-mcp-server/pkg/mcp"
	"github.com/manusa/podman-mcp-server/pkg/version"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var rootCmd = &cobra.Command{
	Use:   "podman-mcp-server [command] [options]",
	Short: "Podman Model Context Protocol (MCP) server",
	Long: `
Podman Model Context Protocol (MCP) server

  # show this help
  podman-mcp-server -h

  # shows version information
  podman-mcp-server --version

  # start STDIO server
  podman-mcp-server

  # start a SSE server on port 8080
  podman-mcp-server --sse-port 8080

  # start a SSE server on port 8443 with a public HTTPS host of example.com
  podman-mcp-server --sse-port 8443 --sse-base-url https://example.com:8443

  # TODO: add more examples`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println(version.Version)
			return
		}
		mcpServer, err := mcp.NewSever()
		if err != nil {
			panic(err)
		}

		var sseServer *server.SSEServer
		if ssePort := viper.GetInt("sse-port"); ssePort > 0 {
			sseServer = mcpServer.ServeSse(viper.GetString("sse-base-url"))
			if err := sseServer.Start(fmt.Sprintf(":%d", ssePort)); err != nil {
				panic(err)
			}
		}
		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
		if sseServer != nil {
			_ = sseServer.Shutdown(cmd.Context())
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	rootCmd.Flags().IntP("sse-port", "", 0, "Start a SSE server on the specified port")
	rootCmd.Flags().StringP("sse-base-url", "", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
