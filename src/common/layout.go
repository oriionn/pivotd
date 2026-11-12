package common

type Rect struct {
	Name       string
	X, Y, W, H int
}

type NewSize struct {
	Name string
	W, H int
}

func Relayout(old []Rect, sizes []NewSize) []Rect {
	out := make([]Rect, len(old))
	copy(out, old)

	for _, s := range sizes {
		idx := -1
		for i, r := range out {
			if r.Name == s.Name {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}

		cur := out[idx]
		dw, dh := s.W-cur.W, s.H-cur.H
		right, bottom := cur.X+cur.W, cur.Y+cur.H

		for i := range out {
			if i == idx {
				continue
			}
			if out[i].X >= right {
				out[i].X += dw
			}
			if out[i].Y >= bottom {
				out[i].Y += dh
			}
		}
		out[idx].W, out[idx].H = s.W, s.H
	}
	return out
}
