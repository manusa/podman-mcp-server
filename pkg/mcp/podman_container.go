package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"strconv"
	"strings"
)

func (s *Server) initPodmanContainer() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("container_inspect",
			mcp.WithDescription("Displays the low-level information and configuration of a Docker or Podman container with the specified container ID or name"),
			mcp.WithString("name", mcp.Description("Docker or Podman container ID or name to displays the information"), mcp.Required()),
		), s.containerInspect},
		{mcp.NewTool("container_list",
			mcp.WithDescription("Prints out information about the running Docker or Podman containers"),
		), s.containerList},
		{mcp.NewTool("container_logs",
			mcp.WithDescription("Displays the logs of a Docker or Podman container with the specified container ID or name"),
			mcp.WithString("name", mcp.Description("Docker or Podman container ID or name to displays the logs"), mcp.Required()),
		), s.containerLogs},
		{mcp.NewTool("container_remove",
			mcp.WithDescription("Removes a Docker or Podman container with the specified container ID or name (rm)"),
			mcp.WithString("name", mcp.Description("Docker or Podman container ID or name to remove"), mcp.Required()),
		), s.containerRemove},
		{mcp.NewTool("container_run",
			mcp.WithDescription("Runs a Docker or Podman container with the specified image name"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to pull"), mcp.Required()),
			mcp.WithArray("ports", mcp.Description("Port mappings to expose on the host. "+
				"Format: <hostPort>:<containerPort>. "+
				"Example: 8080:80. "+
				"(Optional, add only to expose ports)"),
				// TODO: manual fix to ensure that the items property gets initialized (Gemini)
				// https://www.googlecloudcommunity.com/gc/AI-ML/Gemini-API-400-Bad-Request-Array-fields-breaks-function-calling/m-p/769835?nobounce
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
			mcp.WithArray("environment", mcp.Description("Environment variables to set in the container. "+
				"Format: <key>=<value>. "+
				"Example: FOO=bar. "+
				"(Optional, add only to set environment variables)"),
				// TODO: manual fix to ensure that the items property gets initialized (Gemini)
				// https://www.googlecloudcommunity.com/gc/AI-ML/Gemini-API-400-Bad-Request-Array-fields-breaks-function-calling/m-p/769835?nobounce
				func(schema map[string]interface{}) {
					schema["type"] = "array"
					schema["items"] = map[string]interface{}{
						"type": "string",
					}
				},
			),
		), s.containerRun},
		{mcp.NewTool("container_stop",
			mcp.WithDescription("Stops a Docker or Podman running container with the specified container ID or name"),
			mcp.WithString("name", mcp.Description("Docker or Podman container ID or name to stop"), mcp.Required()),
		), s.containerStop},
	}
}

func (s *Server) containerInspect(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerInspect(ctr.Params.Arguments["name"].(string))), nil
}

func (s *Server) containerList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerList()), nil
}

func (s *Server) containerLogs(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerLogs(ctr.Params.Arguments["name"].(string))), nil
}

func (s *Server) containerRemove(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerRemove(ctr.Params.Arguments["name"].(string))), nil
}

func (s *Server) containerRun(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ports := ctr.Params.Arguments["ports"]
	portMappings := make(map[int]int)
	if _, ok := ports.([]interface{}); ok {
		for _, port := range ports.([]interface{}) {
			if _, ok := port.(string); !ok {
				continue
			}
			hostPort, _ := strconv.Atoi(strings.Split(port.(string), ":")[0])
			containerPort, _ := strconv.Atoi(strings.Split(port.(string), ":")[1])
			if hostPort > 0 && containerPort > 0 {
				portMappings[hostPort] = containerPort
			}
		}
	}
	environment := ctr.Params.Arguments["environment"]
	envVariables := make([]string, 0)
	if _, ok := environment.([]interface{}); ok && len(environment.([]interface{})) > 0 {
		for _, env := range environment.([]interface{}) {
			if _, ok = env.(string); !ok {
				continue
			}
			envVariables = append(envVariables, env.(string))
		}
	}
	return NewTextResult(s.podman.ContainerRun(ctr.Params.Arguments["imageName"].(string), portMappings, envVariables)), nil
}

func (s *Server) containerStop(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerStop(ctr.Params.Arguments["name"].(string))), nil
}
