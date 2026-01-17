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

  # start HTTP server on port 8080 (Streamable HTTP at /mcp and SSE at /sse)
  podman-mcp-server --port 8080`,
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
		if port := viper.GetInt("port"); port > 0 {
			// Modern HTTP mode: serve both Streamable HTTP and SSE endpoints
			mux := http.NewServeMux()
			mux.Handle("/mcp", mcpServer.ServeStreamableHTTP())
			mux.Handle("/sse", mcpServer.ServeSse())
			httpServer = &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: mux,
			}
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}
			}()
		} else if ssePort := viper.GetInt("sse-port"); ssePort > 0 {
			// Legacy SSE-only mode for backwards compatibility
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
	rootCmd.Flags().IntP("port", "p", 0, "Start HTTP server on the specified port (Streamable HTTP at /mcp and SSE at /sse)")
	rootCmd.Flags().IntP("sse-port", "", 0, "Start a legacy SSE-only server on the specified port")
	rootCmd.Flags().StringP("sse-base-url", "", "", "SSE public base URL to use when sending the endpoint message (e.g. https://example.com)")
	_ = rootCmd.Flags().MarkDeprecated("sse-port", "use --port instead")
	_ = rootCmd.Flags().MarkDeprecated("sse-base-url", "use --port instead")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
