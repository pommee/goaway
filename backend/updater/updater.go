package updater

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type sendSSE func(string)

func SelfUpdate(sse sendSSE, binaryPath string) error {
	sse("[info] Loading update script")
	scriptPath := "./updater.sh"

	sse("[info] Executing update script")
	cmd := exec.Command("bash", scriptPath, binaryPath)

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

	done := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Contains(text, "Stopping") {
				close(done)
				return
			}
			sse(text)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			sse(scanner.Text())
		}
	}()

	select {
	case <-done:
		return nil
	case err := <-waitCmd(cmd):
		if err != nil {
			return fmt.Errorf("update failed: %v", err)
		}
	}

	return nil
}

func waitCmd(cmd *exec.Cmd) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- cmd.Wait()
	}()
	return ch
}
