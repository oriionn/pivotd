package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"git.oriondev.fr/orion/pivotd/src/common"
	"git.oriondev.fr/orion/pivotd/src/hyprland"
	"git.oriondev.fr/orion/pivotd/src/kde"
)

type Backend int

const (
	BackendHyprland Backend = iota
	BackendKde
)

func (b Backend) String() string {
	switch b {
	case BackendHyprland:
		return "hyprland"
	case BackendKde:
		return "kde"
	default:
		return "unknown"
	}
}

func DetectBackend() (Backend, error) {
	desk := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))

	switch {
	case strings.Contains(desk, "kde") || os.Getenv("KDE_FULL_SESSION") == "true":
		return BackendKde, nil
	case strings.Contains(desk, "hyprland"):
		return BackendHyprland, nil
	case os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "":
		return BackendHyprland, nil
	default:
		return 0, errors.New("unsupported desktop: only Hyprland and KDE are supported")
	}
}


func rotate(name string, rotation common.Rotation) error {
	backend, err := DetectBackend()
	if err != nil { return err }

	switch (backend) {
		case BackendKde:
			return kde.Rotate(name, rotation)
		case BackendHyprland:
			return hyprland.Rotate(name, rotation)
		default:
			return fmt.Errorf("impossible rotation: incompatible backend")
	}
}
