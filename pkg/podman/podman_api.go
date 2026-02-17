//go:build !exclude_podman_api

package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/containers/podman/v5/pkg/bindings"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	"github.com/containers/podman/v5/pkg/bindings/images"
	"github.com/containers/podman/v5/pkg/bindings/network"
	"github.com/containers/podman/v5/pkg/bindings/volumes"
	entitiesTypes "github.com/containers/podman/v5/pkg/domain/entities/types"
	netTypes "go.podman.io/common/libnetwork/types"

	"github.com/manusa/podman-mcp-server/pkg/config"
)

func init() {
	Register(&podmanApi{})
}

// podmanApi implements the Podman interface using the Podman REST API
// via pkg/bindings.
type podmanApi struct {
	ctx          context.Context // Context with connection info
	outputFormat string
	initOnce     sync.Once
	initErr      error
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
func (p *podmanApi) Initialize(cfg config.Config) (Podman, error) {
	instance := &podmanApi{
		outputFormat: cfg.OutputFormat,
	}
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
	if p.outputFormat == config.OutputFormatJSON {
		return toJSON(data)
	}
	return formatContainerList(data), nil
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

	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(p.ctx, 30*time.Second)
	defer cancel()

	// Collect logs in goroutine.
	var stdoutBuf, stderrBuf bytes.Buffer
	done := make(chan struct{})
	go func() {
		defer close(done)
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
			case <-ctx.Done():
				return
			}
			if stdoutChan == nil && stderrChan == nil {
				return
			}
		}
	}()

	// containers.Logs should close both channels when done
	err := containers.Logs(ctx, name, opts, stdoutChan, stderrChan)

	// Wait for collection to finish or context to be cancelled
	select {
	case <-done:
	case <-ctx.Done():
		return "", fmt.Errorf("timeout waiting for container logs")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}

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
	if p.outputFormat == config.OutputFormatJSON {
		return toJSON(data)
	}
	return formatImageList(data), nil
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
	if p.outputFormat == config.OutputFormatJSON {
		return toJSON(data)
	}
	return formatNetworkList(data), nil
}

// VolumeList lists all volumes on the system.
func (p *podmanApi) VolumeList() (string, error) {
	data, err := volumes.List(p.ctx, nil)
	if err != nil {
		return "", err
	}
	if p.outputFormat == config.OutputFormatJSON {
		return toJSON(data)
	}
	return formatVolumeList(data), nil
}

// toJSON converts a value to an indented JSON string.
func toJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatContainerList formats container list data as a text table.
func formatContainerList(data []entitiesTypes.ListContainer) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tPORTS\tNAMES")
	for _, c := range data {
		id := c.ID
		if len(id) > 12 {
			id = id[:12]
		}
		command := ""
		if len(c.Command) > 0 {
			command = strings.Join(c.Command, " ")
			if len(command) > 20 {
				command = command[:20] + "..."
			}
		}
		created := formatTimeAgo(c.Created)
		status := c.Status
		if status == "" {
			status = c.State
		}
		ports := formatPorts(c.Ports)
		names := strings.Join(c.Names, ",")
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			id, c.Image, command, created, status, ports, names)
	}
	_ = w.Flush()
	return strings.TrimSuffix(buf.String(), "\n")
}

// formatImageList formats image list data as a text table.
func formatImageList(data []*entitiesTypes.ImageSummary) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "REPOSITORY\tTAG\tDIGEST\tIMAGE ID\tCREATED\tSIZE")
	for _, img := range data {
		repo := "<none>"
		tag := "<none>"
		if len(img.RepoTags) > 0 {
			parts := strings.Split(img.RepoTags[0], ":")
			if len(parts) >= 2 {
				repo = strings.Join(parts[:len(parts)-1], ":")
				tag = parts[len(parts)-1]
			} else {
				repo = img.RepoTags[0]
			}
		} else if len(img.Names) > 0 {
			parts := strings.Split(img.Names[0], ":")
			if len(parts) >= 2 {
				repo = strings.Join(parts[:len(parts)-1], ":")
				tag = parts[len(parts)-1]
			} else {
				repo = img.Names[0]
			}
		}
		id := strings.TrimPrefix(img.ID, "sha256:")
		if len(id) > 12 {
			id = id[:12]
		}
		digest := img.Digest
		if len(digest) > 19 {
			digest = digest[:19]
		}
		if digest == "" {
			digest = "<none>"
		}
		created := formatTimeAgo(time.Unix(img.Created, 0))
		size := formatSize(img.Size)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			repo, tag, digest, id, created, size)
	}
	_ = w.Flush()
	return strings.TrimSuffix(buf.String(), "\n")
}

// formatNetworkList formats network list data as a text table.
func formatNetworkList(data []netTypes.Network) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER")
	for _, n := range data {
		id := n.ID
		if len(id) > 12 {
			id = id[:12]
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", id, n.Name, n.Driver)
	}
	_ = w.Flush()
	return strings.TrimSuffix(buf.String(), "\n")
}

// formatVolumeList formats volume list data as a text table.
func formatVolumeList(data []*entitiesTypes.VolumeListReport) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "DRIVER\tVOLUME NAME")
	for _, v := range data {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", v.Driver, v.Name)
	}
	_ = w.Flush()
	return strings.TrimSuffix(buf.String(), "\n")
}

// formatTimeAgo formats a time as a human-readable "X ago" string.
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}

// formatSize formats a size in bytes as a human-readable string.
func formatSize(size int64) string {
	const (
		KB = 1000
		MB = 1000 * KB
		GB = 1000 * MB
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// formatPorts formats port mappings as a string.
func formatPorts(ports []netTypes.PortMapping) string {
	if len(ports) == 0 {
		return ""
	}
	var parts []string
	for _, p := range ports {
		if p.HostPort > 0 {
			parts = append(parts, fmt.Sprintf("%d->%d/%s", p.HostPort, p.ContainerPort, p.Protocol))
		} else {
			parts = append(parts, fmt.Sprintf("%d/%s", p.ContainerPort, p.Protocol))
		}
	}
	return strings.Join(parts, ", ")
}

// Ensure interface compliance at compile time.
var _ Podman = (*podmanApi)(nil)
