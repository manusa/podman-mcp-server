package podman

import (
	"sort"
	"strings"

	"github.com/manusa/podman-mcp-server/pkg/config"
)

// Podman interface
type Podman interface {
	// ContainerInspect displays the low-level information on containers identified by the ID or name
	ContainerInspect(name string) (string, error)
	// ContainerList lists all the containers on the system
	ContainerList() (string, error)
	// ContainerLogs Display the logs of a container
	ContainerLogs(name string) (string, error)
	// ContainerRemove removes a container
	ContainerRemove(name string) (string, error)
	// ContainerRun pulls an image from a registry
	ContainerRun(imageName string, portMappings map[int]int, envVariables []string) (string, error)
	// ContainerStop stops a running container using the ID or name
	ContainerStop(name string) (string, error)
	// ImageBuild builds an image from a Dockerfile, Podmanfile, or Containerfile
	ImageBuild(containerFile string, imageName string) (string, error)
	// ImageList list the container images on the system
	ImageList() (string, error)
	// ImagePull pulls an image from a registry
	ImagePull(imageName string) (string, error)
	// ImagePush pushes an image to a registry
	ImagePush(imageName string) (string, error)
	// ImageRemove removes an image from the system
	ImageRemove(imageName string) (string, error)
	// NetworkList lists all the networks on the system
	NetworkList() (string, error)
	// VolumeList lists all the volumes on the system
	VolumeList() (string, error)
}

// NewPodman returns a Podman implementation.
// If cfg.PodmanImpl is empty, auto-detects by iterating implementations by priority.
// If cfg.PodmanImpl is specified, returns that implementation or error if unavailable.
func NewPodman(cfg config.Config) (Podman, error) {
	if cfg.PodmanImpl != "" {
		// User specified an implementation
		impl := ImplementationFromString(cfg.PodmanImpl)
		if impl == nil {
			return nil, &ErrUnknownImplementation{Name: cfg.PodmanImpl, Available: ImplementationNames()}
		}
		if !impl.Available() {
			return nil, &ErrImplementationNotAvailable{Name: cfg.PodmanImpl, Reason: "not available on this system"}
		}
		return impl.Initialize(cfg)
	}

	// Auto-detect: sort by priority (descending) and return first available
	impls := Implementations()
	sort.Slice(impls, func(i, j int) bool {
		return impls[i].Priority() > impls[j].Priority()
	})

	var tried []string
	for _, impl := range impls {
		if impl.Available() {
			return impl.Initialize(cfg)
		}
		tried = append(tried, impl.Name()+" (not available)")
	}

	return nil, &ErrNoImplementationAvailable{TriedImplementations: tried}
}

// ErrUnknownImplementation is returned when an invalid implementation is specified.
type ErrUnknownImplementation struct {
	Name      string
	Available []string
}

func (e *ErrUnknownImplementation) Error() string {
	return "invalid podman implementation \"" + e.Name + "\", valid options: " + strings.Join(e.Available, ", ")
}
