package updater

import (
	"bufio"
	"fmt"
	"os/exec"
)

type sendSSE func(string)

func SelfUpdate(sse sendSSE) error {
	sse("[INFO] Loading update script")
	scriptPath := "./updater.sh"

	sse("[INFO] Executing update script")
	cmd := exec.Command("bash", scriptPath)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			sse(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			sse(scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("update failed: %v", err)
	}

	return nil
}
