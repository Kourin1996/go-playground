package pointer

func ToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func ToByte(b *byte) byte {
	if b == nil {
		return 0
	}
	return *b
}

func ToComplex64(c *complex64) complex64 {
	if c == nil {
		return 0
	}
	return *c
}

func ToComplex128(c *complex128) complex128 {
	if c == nil {
		return 0
	}
	return *c
}

func ToError(e *error) error {
	if e == nil {
		return nil
	}
	return *e
}

func ToFloat32(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

func ToFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func ToInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func ToInt8(i *int8) int8 {
	if i == nil {
		return 0
	}
	return *i
}

func ToInt16(i *int16) int16 {
	if i == nil {
		return 0
	}
	return *i
}

func ToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func ToInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func ToRune(r *rune) rune {
	if r == nil {
		return 0
	}
	return *r
}

func ToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ToUint(u *uint) uint {
	if u == nil {
		return 0
	}
	return *u
}

func ToUint8(u *uint8) uint8 {
	if u == nil {
		return 0
	}
	return *u
}

func ToUint16(u *uint16) uint16 {
	if u == nil {
		return 0
	}
	return *u
}

func ToUint32(u *uint32) uint32 {
	if u == nil {
		return 0
	}
	return *u
}

func ToUint64(u *uint64) uint64 {
	if u == nil {
		return 0
	}
	return *u
}
