package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/manusa/podman-mcp-server/pkg/mcp"
	"github.com/manusa/podman-mcp-server/pkg/version"
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

		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		// Handle signals for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		mcpServer, err := mcp.NewServer()
		if err != nil {
			panic(err)
		}

		var httpServer *http.Server
		if ssePort := viper.GetInt("sse-port"); ssePort > 0 {
			sseHandler := mcpServer.ServeSse()
			httpServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", ssePort),
				Handler: sseHandler,
			}
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
			}()
		}

		if err := mcpServer.ServeStdio(ctx); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}

		if httpServer != nil {
			_ = httpServer.Shutdown(context.Background())
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
