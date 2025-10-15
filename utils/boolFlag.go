package utils

import (
	"fmt"
	"strconv"
)

type BoolFlag struct {
	IsSet bool
	Value bool
}

func (b *BoolFlag) String() string {
	return fmt.Sprintf("%v", b.Value)
}

func (b *BoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("invalid boolean: %s", s)
	}
	b.IsSet = true
	b.Value = v
	return nil
}

func (b *BoolFlag) Type() string {
	return "bool"
}
