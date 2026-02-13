//go:build !exclude_podman_api

package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/bindings/network"
	"github.com/containers/podman/v5/pkg/bindings/volumes"

	"github.com/manusa/podman-mcp-server/pkg/config"
)

func init() {
	Register(&podmanApi{})
}

// podmanApi implements the Podman interface using the Podman REST API
// via pkg/bindings.
type podmanApi struct {
	ctx      context.Context // Context with connection info
	initOnce sync.Once
	initErr  error
}

// Name returns the unique identifier for this implementation.
func (p *podmanApi) Name() string {
	return "api"
}

// Description returns a human-readable description for help text.
func (p *podmanApi) Description() string {
	return "Podman REST API via Unix socket"
}

// Available returns true if this implementation can be used.
// It checks if a Podman socket is available and responds to ping.
func (p *podmanApi) Available() bool {
	socketPath, err := DetectSocket()
	if err != nil {
		return false
	}
	return PingSocket(socketPath) == nil
}

// Priority returns the priority for auto-detection.
// API has priority 100 (higher than CLI which has 50).
func (p *podmanApi) Priority() int {
	return 100
}

// Initialize creates and initializes a new podmanApi instance.
func (p *podmanApi) Initialize(_ config.Config) (Podman, error) {
	instance := &podmanApi{}
	if err := instance.ensureConnection(); err != nil {
		return nil, err
	}
	return instance, nil
}

// ensureConnection establishes a connection to the Podman socket.
// The connection is established once and reused for all operations.
func (p *podmanApi) ensureConnection() error {
	p.initOnce.Do(func() {
		socketPath, err := DetectSocket()
		if err != nil {
			p.initErr = fmt.Errorf("failed to detect socket: %w", err)
			return
		}
		p.ctx, p.initErr = bindings.NewConnection(context.Background(), socketPath)
		if p.initErr != nil {
			p.initErr = fmt.Errorf("failed to connect to socket: %w", p.initErr)
		}
	})
	return p.initErr
}

// ContainerInspect displays the low-level information on containers identified by ID or name.
func (p *podmanApi) ContainerInspect(name string) (string, error) {
	data, err := containers.Inspect(p.ctx, name, nil)
	if err != nil {
		return "", err
	}
	return toJSON(data)
}

// ContainerList lists all containers on the system.
func (p *podmanApi) ContainerList() (string, error) {
	all := true
	opts := &containers.ListOptions{
		All: &all,
	}
	data, err := containers.List(p.ctx, opts)
	if err != nil {
		return "", err
	}
	return toJSON(data)
}

// ContainerLogs returns the logs of a container.
func (p *podmanApi) ContainerLogs(name string) (string, error) {
	stdout := true
	stderr := true
	opts := &containers.LogOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	stdoutChan := make(chan string)
	stderrChan := make(chan string)

	// Collect logs in goroutine
	var stdoutBuf, stderrBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		for {
			select {
			case line, ok := <-stdoutChan:
				if !ok {
					stdoutChan = nil
				} else {
					stdoutBuf.WriteString(line)
				}
			case line, ok := <-stderrChan:
				if !ok {
					stderrChan = nil
				} else {
					stderrBuf.WriteString(line)
				}
			}
			if stdoutChan == nil && stderrChan == nil {
				close(done)
				return
			}
		}
	}()

	err := containers.Logs(p.ctx, name, opts, stdoutChan, stderrChan)
	if err != nil {
		return "", err
	}

	// Wait for collection to finish
	<-done

	// Combine stdout and stderr
	result := stdoutBuf.String()
	if stderrContent := stderrBuf.String(); stderrContent != "" {
		if result != "" {
			result += "\n"
		}
		result += stderrContent
	}

	return result, nil
}

// ContainerRemove removes a container.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ContainerRemove(_ string) (string, error) {
	return "", fmt.Errorf("ContainerRemove not yet implemented for API implementation")
}

// ContainerRun runs a new container from an image.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ContainerRun(_ string, _ map[int]int, _ []string) (string, error) {
	return "", fmt.Errorf("ContainerRun not yet implemented for API implementation")
}

// ContainerStop stops a running container.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ContainerStop(_ string) (string, error) {
	return "", fmt.Errorf("ContainerStop not yet implemented for API implementation")
}

// ImageBuild builds an image from a Containerfile.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ImageBuild(_ string, _ string) (string, error) {
	return "", fmt.Errorf("ImageBuild not yet implemented for API implementation")
}

// ImageList lists all images on the system.
func (p *podmanApi) ImageList() (string, error) {
	all := true
	opts := &images.ListOptions{
		All: &all,
	}
	data, err := images.List(p.ctx, opts)
	if err != nil {
		return "", err
	}
	return toJSON(data)
}

// ImagePull pulls an image from a registry.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ImagePull(_ string) (string, error) {
	return "", fmt.Errorf("ImagePull not yet implemented for API implementation")
}

// ImagePush pushes an image to a registry.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ImagePush(_ string) (string, error) {
	return "", fmt.Errorf("ImagePush not yet implemented for API implementation")
}

// ImageRemove removes an image from the system.
// Note: This is a write operation but included as stub for interface compliance.
// Full implementation will be added in Phase 3.
func (p *podmanApi) ImageRemove(_ string) (string, error) {
	return "", fmt.Errorf("ImageRemove not yet implemented for API implementation")
}

// NetworkList lists all networks on the system.
func (p *podmanApi) NetworkList() (string, error) {
	data, err := network.List(p.ctx, nil)
	if err != nil {
		return "", err
	}
	return toJSON(data)
}

// VolumeList lists all volumes on the system.
func (p *podmanApi) VolumeList() (string, error) {
	data, err := volumes.List(p.ctx, nil)
	if err != nil {
		return "", err
	}
	return toJSON(data)
}

// toJSON converts a value to an indented JSON string.
func toJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Ensure interface compliance at compile time.
var _ Podman = (*podmanApi)(nil)
