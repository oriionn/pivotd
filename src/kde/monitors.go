package kde

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type mode struct {
	ID   json.RawMessage `json:"id"`
	Size size            `json:"size"`
}

type output struct {
	Name          string          `json:"name"`
	Enabled       bool            `json:"enabled"`
	Connected     bool            `json:"connected"`
	CurrentModeID json.RawMessage `json:"currentModeId"`
	Modes         []mode          `json:"modes"`
	Size          size            `json:"size"`
	Scale         float64         `json:"scale"`
	Rotation      int             `json:"rotation"`
	Pos           struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"pos"`
}

type config struct {
	Outputs []output `json:"outputs"`
}

func outputs() ([]output, error) {
	raw, err := exec.Command("kscreen-doctor", "-j").Output()
	if err != nil {
		return nil, fmt.Errorf("kscreen-doctor -j: %w", err)
	}

	var cfg config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parsing kscreen-doctor output: %w", err)
	}
	if cfg.Outputs == nil {
		return nil, fmt.Errorf("no outputs")
	}

	return cfg.Outputs, nil
}
