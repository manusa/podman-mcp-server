package cmd

import (
	"errors"
	"fmt"
	"github.com/manusa/podman-mcp-server/pkg/mcp"
	"github.com/manusa/podman-mcp-server/pkg/version"
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
		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
