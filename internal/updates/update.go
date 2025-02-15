package update

import (
	"encoding/json"
	"fmt"
	"goaway/internal/logging"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

var log = logging.GetLogger()

const (
	maxRetries    = 10
	retryDelay    = 3 * time.Second
	checkInterval = time.Hour
	releaseURL    = "https://github.com/pommee/goaway/releases/latest/download/goaway-linux-amd64"
)

type TagsResponse struct {
	Name       string `json:"name"`
	ZipballURL string `json:"zipball_url"`
	TarballURL string `json:"tarball_url"`
	Commit     struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	NodeID string `json:"node_id"`
}

func PollUpdates(currentVersion string) {
	for {
		nextCheck := time.Now().Add(checkInterval).Format("15:04:05")
		log.Debug("Next update check scheduled for %s", nextCheck)

		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				log.Debug("[Retry %d] Checking for new versions...", retry)
			} else {
				log.Debug("Checking for new versions...")
			}

			if err := checkForUpdate(currentVersion); err != nil {
				log.Warning("Update check failed: %v", err)
				time.Sleep(retryDelay)
				continue
			}
			break
		}

		time.Sleep(checkInterval)
	}
}

func checkForUpdate(currentVersion string) error {
	tags, err := getTags()
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %v", err)
	}

	for _, tag := range tags {
		if isNewVersion(currentVersion, tag.Name) {
			log.Info("New version available: %s (Current: %s)", tag.Name, currentVersion)
			return downloadAndReplaceBinary()
		}
	}
	log.Debug("No new updates available.")
	return nil
}

func isNewVersion(current, latest string) bool {
	return strings.TrimPrefix(latest, "v") > strings.TrimPrefix(current, "v")
}

func getTags() ([]TagsResponse, error) {
	const url = "https://api.github.com/repos/pommee/goaway/tags"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed: %s", resp.Status)
	}

	var tags []TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tags, nil
}

func downloadAndReplaceBinary() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable path: %w", err)
	}

	tempFile := execPath + ".new"

	log.Info("Downloading new binary from %s...", releaseURL)
	resp, err := http.Get(releaseURL)
	if err != nil {
		return fmt.Errorf("failed to download new binary: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save new binary: %w", err)
	}

	if err := os.Chmod(tempFile, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	log.Info("Replacing binary at %s", execPath)
	if err := os.Rename(tempFile, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return restartProcess()
}

func restartProcess() error {
	log.Info("Restarting process...")

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable path: %w", err)
	}

	args := os.Args
	env := os.Environ()
	if err := syscall.Exec(execPath, args, env); err != nil {
		return fmt.Errorf("failed to restart process: %w", err)
	}

	return nil
}
