package podman

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type podmanCli struct {
	filePath string
}

// ContainerInspect
// https://docs.podman.io/en/stable/markdown/podman-inspect.1.html
func (p *podmanCli) ContainerInspect(name string) (string, error) {
	return p.exec("inspect", name)
}

// ContainerList
// https://docs.podman.io/en/stable/markdown/podman-ps.1.html
func (p *podmanCli) ContainerList() (string, error) {
	return p.exec("container", "list", "-a")
}

// ContainerLogs
// https://docs.podman.io/en/stable/markdown/podman-logs.1.html
func (p *podmanCli) ContainerLogs(name string) (string, error) {
	return p.exec("logs", name)
}

// ContainerRemove
// https://docs.podman.io/en/stable/markdown/podman-rm.1.html
func (p *podmanCli) ContainerRemove(name string) (string, error) {
	return p.exec("container", "rm", name)
}

// ContainerRun
// https://docs.podman.io/en/stable/markdown/podman-run.1.html
func (p *podmanCli) ContainerRun(imageName string, portMappings map[int]int, envVariables []string) (string, error) {
	args := []string{"run", "--rm", "-d"}
	if len(portMappings) > 0 {
		for hostPort, containerPort := range portMappings {
			args = append(args, fmt.Sprintf("--publish=%d:%d", hostPort, containerPort))
		}
	} else {
		args = append(args, "--publish-all")
	}
	for _, env := range envVariables {
		args = append(args, "--env", env)
	}
	output, err := p.exec(append(args, imageName)...)
	if err == nil {
		return output, nil
	}
	if strings.Contains(output, "Error: short-name") {
		imageName = "docker.io/" + imageName
		if output, err = p.exec(append(args, imageName)...); err == nil {
			return output, nil
		}
	}
	return "", err
}

// ContainerStop
// https://docs.podman.io/en/stable/markdown/podman-stop.1.html
func (p *podmanCli) ContainerStop(name string) (string, error) {
	return p.exec("container", "stop", name)
}

// ImageBuild
// https://docs.podman.io/en/stable/markdown/podman-build.1.html
func (p *podmanCli) ImageBuild(containerFile string, imageName string) (string, error) {
	args := []string{"build"}
	if imageName != "" {
		args = append(args, "-t", imageName)
	}
	return p.exec(append(args, containerFile)...)
}

// ImageList
// https://docs.podman.io/en/stable/markdown/podman-images.1.html
func (p *podmanCli) ImageList() (string, error) {
	return p.exec("images", "--digests")
}

// ImagePull
// https://docs.podman.io/en/stable/markdown/podman-pull.1.html
func (p *podmanCli) ImagePull(imageName string) (string, error) {
	output, err := p.exec("image", "pull", imageName)
	if err == nil {
		return fmt.Sprintf("%s\n%s pulled successfully", output, imageName), nil
	}
	if strings.Contains(output, "Error: short-name") {
		imageName = "docker.io/" + imageName
		if output, err = p.exec("pull", imageName); err == nil {
			return fmt.Sprintf("%s\n%s pulled successfully", output, imageName), nil
		}
	}
	return "", err
}

// ImagePush
// https://docs.podman.io/en/stable/markdown/podman-push.1.html
func (p *podmanCli) ImagePush(imageName string) (string, error) {
	output, err := p.exec("image", "push", imageName)
	if err == nil {
		return fmt.Sprintf("%s\n%s pushed successfully", output, imageName), nil
	}
	return "", err
}

// ImageRemove
// https://docs.podman.io/en/stable/markdown/podman-rmi.1.html
func (p *podmanCli) ImageRemove(imageName string) (string, error) {
	return p.exec("image", "rm", imageName)
}

// NetworkList
// https://docs.podman.io/en/stable/markdown/podman-network-ls.1.html
func (p *podmanCli) NetworkList() (string, error) {
	return p.exec("network", "ls")
}

// VolumeList
// https://docs.podman.io/en/stable/markdown/podman-volume-ls.1.html
func (p *podmanCli) VolumeList() (string, error) {
	return p.exec("volume", "ls")
}

func (p *podmanCli) exec(args ...string) (string, error) {
	output, err := exec.Command(p.filePath, args...).CombinedOutput()
	return string(output), err
}

func newPodmanCli() (*podmanCli, error) {
	for _, cmd := range []string{"podman", "podman.exe"} {
		filePath, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}
		if _, err = exec.Command(filePath, "version").CombinedOutput(); err == nil {
			return &podmanCli{filePath}, nil
		}
	}
	return nil, errors.New("podman CLI not found")
}
