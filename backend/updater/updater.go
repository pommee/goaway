package updater

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type sendSSE func(string)

func SelfUpdate(sse sendSSE, binaryPath string) error {
	sse("[info] Loading update script")
	scriptPath := "./updater.sh"

	_, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		sse("[info] Script not found locally, downloading from GitHub")
		scriptURL := "https://raw.githubusercontent.com/pommee/goaway/refs/heads/main/updater.sh"

		resp, err := http.Get(scriptURL)
		if err != nil {
			return fmt.Errorf("failed to download script: %w", err)
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download script: HTTP %d", resp.StatusCode)
		}

		scriptContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read script content: %w", err)
		}

		if err := os.WriteFile(scriptPath, scriptContent, 0755); err != nil {
			return fmt.Errorf("failed to write script file: %w", err)
		}
		sse("[info] Script downloaded successfully")
	}

	// Determine which shell to use
	shell := "bash"
	if _, err := exec.LookPath("bash"); err != nil {
		sse("[info] bash not found, falling back to ash")
		shell = "ash"
		if _, err := exec.LookPath("ash"); err != nil {
			return fmt.Errorf("neither bash nor ash found in PATH")
		}
	}

	sse(fmt.Sprintf("[info] Executing update script with %s", shell))
	cmd := exec.Command(shell, scriptPath, binaryPath)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
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
			return fmt.Errorf("update failed: %w", err)
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
