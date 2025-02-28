package podman

import (
	"errors"
	"os/exec"
	"strings"
)

type podmanCli struct {
	filePath string
}

func (p *podmanCli) ContainerList() (string, error) {
	return p.exec("container", "list", "-a")
}

func (p *podmanCli) ContainerRun(imageName string) (string, error) {
	args := []string{"run", "--rm", "-d"}
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

func (p *podmanCli) ContainerStop(name string) (string, error) {
	return p.exec("container", "stop", name)
}

func (p *podmanCli) ImagePull(imageName string) (string, error) {
	output, err := p.exec("pull", imageName)
	if err == nil {
		return output, nil
	}
	if strings.Contains(output, "Error: short-name") {
		imageName = "docker.io/" + imageName
		if output, err = p.exec("pull", imageName); err == nil {
			return output, nil
		}
	}
	return "", err
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
