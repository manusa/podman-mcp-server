package podman

import "strings"

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
// If override is empty or not provided, auto-detects by using the default implementation.
// If override is specified, returns that implementation or error if unavailable.
// Currently supported implementations:
//   - "cli" (default): Uses podman/docker CLI
//   - Future: "api" will use Podman REST API via Unix socket
func NewPodman(override ...string) (Podman, error) {
	impl := ""
	if len(override) > 0 {
		impl = override[0]
	}
	// TODO: implement registry pattern with multiple implementations (Phase 1)
	// For now, only CLI is supported
	switch impl {
	case "", "cli":
		return newPodmanCli()
	default:
		return nil, &ErrUnknownImplementation{Name: impl, Available: []string{"cli"}}
	}
}

// ErrUnknownImplementation is returned when an invalid implementation is specified.
type ErrUnknownImplementation struct {
	Name      string
	Available []string
}

func (e *ErrUnknownImplementation) Error() string {
	return "invalid podman implementation \"" + e.Name + "\", valid options: " + strings.Join(e.Available, ", ")
}
