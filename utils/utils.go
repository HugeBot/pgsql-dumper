package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func GetContainerId(containerCLI string, contaierPattern string) (string, error) {
	var containerId string

	if contaierPattern != "" {
		cmd := exec.Command(containerCLI, "ps", "-afq", fmt.Sprintf("'name=^%s'", contaierPattern))
		cmd.Stderr = os.Stderr

		out, err := cmd.StdoutPipe()
		if err != nil {
			return containerId, err
		}

		if err := cmd.Start(); err != nil {
			return containerId, err
		}

		byteArray, err := io.ReadAll(out)
		if err != nil {
			return containerId, err
		}

		containerId = string(byteArray)

		if containerId == "" {
			return containerId, fmt.Errorf("container pattern %s not correspond to existing container", containerId)
		}

		return containerId, nil
	} else {
		return containerId, errors.New("containerPattern is empty")
	}
}
