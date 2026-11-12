package hyprland

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Get monitors from hyprctl
func monitors() ([]monitor, error) {
	raw, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl monitors -j: %w", err)
	}

	var list []monitor
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parsing hyprctl monitors output: %w", err)
	}
	return list, nil
}
