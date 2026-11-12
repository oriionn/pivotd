package kde

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git.oriondev.fr/orion/pivotd/src/common"
)

func Rotate(name string, rot common.Rotation) error {
	args := []string{fmt.Sprintf("output.%s.rotation.%s", name, rot.Xrandr())}

	if pos, err := Positions(name, rot); err != nil {
		fmt.Fprintf(os.Stderr, "warning: no re-layout (%v)\n", err)
	} else {
		args = append(args, pos...)
	}

	out, err := exec.Command("kscreen-doctor", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("kscreen-doctor %s: %w: %s",
			strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
