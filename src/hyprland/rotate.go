package hyprland

import (
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"git.oriondev.fr/orion/pivotd/src/common"
)

type monitor struct {
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	X           int     `json:"x"`
	Y           int     `json:"y"`
	Scale       float64 `json:"scale"`
	Transform   int     `json:"transform"`
	Disabled    bool    `json:"disabled"`
}

type decl struct {
	name      string
	mode      string
	x, y      int
	scale     float64
	transform int
}

func Rotate(name string, rot common.Rotation) error {
	list, err := monitors()
	if err != nil {
		return err
	}

	var (
		mons    []decl
		old     []common.Rect
		newSize []common.NewSize
	)

	// For all monitors in the list
	for _, m := range list {
		if m.Disabled { continue }

		// Calculate logical size
		scale := m.Scale
		if scale <= 0 { scale = 1 }
		w := int(math.Round(float64(m.Width) / scale))
		h := int(math.Round(float64(m.Height) / scale))

		cw, ch := w, h
		if common.FromTransform(m.Transform).Quarter() { cw, ch = h, w }

		// Transform data based on the rotation
		t := m.Transform
		if m.Name == name {
			nw, nh := w, h
			if rot.Quarter() { nw, nh = h, w }
			newSize = append(newSize, common.NewSize{Name: m.Name, W: nw, H: nh})
			t = rot.Transform()
		}

		refresh := m.RefreshRate
		if refresh == 0 { refresh = 60 }

		old = append(old, common.Rect{Name: m.Name, X: m.X, Y: m.Y, W: cw, H: ch})
		mons = append(mons, decl{
			name:      m.Name,
			mode:      fmt.Sprintf("%dx%d@%.3f", m.Width, m.Height, refresh),
			scale:     scale,
			transform: t,
		})
	}

	if len(newSize) == 0 {
		return fmt.Errorf("Hyprland has no enabled monitor %s", name)
	}

	for i, r := range common.Relayout(old, newSize) { mons[i].x, mons[i].y = r.X, r.Y }

	var lua strings.Builder
	for _, m := range mons {
		fmt.Fprintf(&lua,
			"hl.monitor({ output = %q, mode = %q, position = \"%dx%d\", scale = %s, transform = %d })\n",
			m.name, m.mode, m.x, m.y,
			strconv.FormatFloat(m.scale, 'f', -1, 64),
			m.transform,
		)
	}

	// Call declarations to Hyprland
	out, err := exec.Command("hyprctl", "eval", lua.String()).Output()
	reply := strings.TrimSpace(string(out))
	if err != nil {
		return fmt.Errorf("hyprctl eval: %w: %s", err, reply)
	}

	for l := range strings.SplitSeq(reply, "\n") {
		if l = strings.TrimSpace(l); l != "" && l != "ok" {
			return fmt.Errorf("hyprctl eval failed: %s", reply)
		}
	}

	return nil
}
