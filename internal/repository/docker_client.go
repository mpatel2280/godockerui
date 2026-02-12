package repository

import "os/exec"

func DockerBinaryPath() (string, error) {
	return exec.LookPath("docker")
}
