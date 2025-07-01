package vm

import (
	"log"
	"os"
	"os/exec"
	"strconv"
)

func StartVM(containerID, appName string, videoPort, audioPort, syncPort int) error {
	log.Printf("Start VM %s for app %s\n", containerID, appName)

	params := []string{
		containerID,
		strconv.Itoa(videoPort),
		strconv.Itoa(audioPort),
		strconv.Itoa(syncPort),
		appName,
	}

	cmd := exec.Command("./startVM.sh", params...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("[%s] Failed to start VM\n", containerID)
		return err
	}

	return nil
}

func StopVM(containerID, appName string) error {
	log.Printf("[%s] Stopping VM\n", containerID)

	params := []string{
		containerID,
		appName,
	}
	cmd := exec.Command("./stopVM.sh", params...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
