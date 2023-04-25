package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func GetContainerId(containerCLI string, contaierPattern string) (string, error) {
	var containerId string

	if contaierPattern != "" {
		cmd := exec.Command(containerCLI, "ps", "-aqf", fmt.Sprintf("\"name=^%s\"", contaierPattern))
		cmd.Stderr = os.Stderr

		out, err := cmd.StdoutPipe()
		if err != nil {
			return containerId, err
		}

		log.Printf("Running command %v\n", cmd.Args)
		if err := cmd.Start(); err != nil {
			return containerId, err
		}

		byteArray, err := io.ReadAll(out)
		if err != nil {
			return containerId, err
		}

		containerId = strings.Split(string(byteArray), "\n")[0]

		if containerId == "" {
			return containerId, fmt.Errorf("container pattern %s not correspond to existing container", contaierPattern)
		}

		return containerId, nil
	} else {
		return containerId, errors.New("containerPattern is empty")
	}
}
