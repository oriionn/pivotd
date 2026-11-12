package common

type Rotation int
const (
	RotationNone     Rotation = 1
	RotationRight    Rotation = 2
	RotationInverted Rotation = 3
	RotationLeft     Rotation = 4
)

func (r Rotation) Quarter() bool {
	return r == RotationRight || r == RotationLeft
}

func (r Rotation) String() string {
	switch r {
		case RotationRight:
			return "Right"
		case RotationInverted:
			return "Inverted"
		case RotationLeft:
			return "Left"
		default:
			return "Normal"
	}
}

func (r Rotation) Xrandr() string {
	switch r {
	case RotationRight:
		return "right"
	case RotationInverted:
		return "inverted"
	case RotationLeft:
		return "left"
	default:
		return "none"
	}
}

func (r Rotation) Transform() int {
	switch r {
	case RotationLeft:
		return 1
	case RotationInverted:
		return 2
	case RotationRight:
		return 3
	default:
		return 0
	}
}

func FromTransform(t int) Rotation {
	switch ((t % 4) + 4) % 4 {
	case 1:
		return RotationLeft
	case 2:
		return RotationInverted
	case 3:
		return RotationRight
	default:
		return RotationNone
	}
}

func FromKscreen(k int) Rotation {
	switch k {
	case 2:
		return RotationLeft
	case 4:
		return RotationInverted
	case 8:
		return RotationRight
	default:
		return RotationNone
	}
}
