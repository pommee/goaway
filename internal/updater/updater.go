package updater

import (
	"fmt"
	"goaway/internal/logging"
	"os"
	"os/exec"
)

var log = logging.GetLogger()

func SelfUpdate() error {
	// Path to the update script
	scriptPath := "./updater.sh"

	// Execute the script
	cmd := exec.Command("sh", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Info("Starting update process...")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("update failed: %v", err)
	}

	log.Info("Update completed successfully.")
	return nil
}
