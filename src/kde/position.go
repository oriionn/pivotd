package kde

import (
	"bytes"
	"fmt"
	"math"

	"git.oriondev.fr/orion/pivotd/src/common"
)

type size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func Positions(name string, rot common.Rotation) ([]string, error) {
	outs, err := outputs()
	if err != nil {
		return nil, err
	}

	var old []common.Rect
	var newSize []common.NewSize

	for _, o := range outs {
		if !o.Enabled || !o.Connected { continue }

		sz := o.Size
		for _, m := range o.Modes {
			if bytes.Equal(m.ID, o.CurrentModeID) {
				sz = m.Size
				break
			}
		}

		scale := o.Scale
		if scale <= 0 { scale = 1 }
		w := int(math.Round(sz.Width / scale))
		h := int(math.Round(sz.Height / scale))

		cw, ch := w, h
		if common.FromKscreen(o.Rotation).Quarter() { cw, ch = h, w }

		if o.Name == name {
			nw, nh := w, h
			if rot.Quarter() { nw, nh = h, w }
			newSize = append(newSize, common.NewSize{Name: o.Name, W: nw, H: nh})
		}

		old = append(old, common.Rect{Name: o.Name, X: o.Pos.X, Y: o.Pos.Y, W: cw, H: ch})
	}

	if len(newSize) == 0 {
		return nil, fmt.Errorf("KDE has no enabled output %s", name)
	}

	updated := common.Relayout(old, newSize)

	var args []string
	for i, n := range updated {
		if n.X != old[i].X || n.Y != old[i].Y {
			args = append(args, fmt.Sprintf("output.%s.position.%d,%d", n.Name, n.X, n.Y))
		}
	}
	return args, nil
}
