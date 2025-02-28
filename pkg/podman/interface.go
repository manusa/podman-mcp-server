package podman

// Podman interface
type Podman interface {
	// ContainerInspect displays the low-level information on containers identified by the ID or name
	ContainerInspect(name string) (string, error)
	// ContainerList lists all the containers on the system
	ContainerList() (string, error)
	// ContainerLogs Display the logs of a container
	ContainerLogs(name string) (string, error)
	// ImagePull pulls an image from a registry
	ImagePull(imageName string) (string, error)
	// ContainerRun pulls an image from a registry
	ContainerRun(imageName string) (string, error)
	// ContainerStop stops a running container using the ID or name
	ContainerStop(name string) (string, error)
}

func NewPodman() (Podman, error) {
	// TODO: add implementations for Podman bindings and Docker CLI
	return newPodmanCli()
}
