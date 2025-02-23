package updater

import (
	"fmt"
	"os"
	"os/exec"
)

func SelfUpdate() error {
	scriptPath := "./updater.sh"

	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Starting update process...")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("update failed: %v", err)
	}

	return nil
}
