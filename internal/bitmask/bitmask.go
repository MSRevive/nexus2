package bitmask

import (
)

type Bitmask uint32

const (
	BANNED Bitmask = 1 << iota
	DONOR
	ADMIN
)

func (f Bitmask) HasFlag(flag Bitmask) bool { 
	return f&flag != 0 
}

func (f *Bitmask) AddFlag(flag Bitmask) {
	*f |= flag 
}

func (f *Bitmask) ClearFlag(flag Bitmask) {
	*f &= ^flag 
}

func (f *Bitmask) ToggleFlag(flag Bitmask) { 
	*f ^= flag 
}