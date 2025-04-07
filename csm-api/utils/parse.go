package utils

import (
	"github.com/guregu/null"
	"strconv"
)

func ParseNullString(s string) null.String {
	if s == "" {
		return null.NewString("", false)
	}
	return null.NewString(s, true)
}

func ParseNullInt(s string) null.Int {
	if s == "" {
		return null.NewInt(0, false)
	}

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return null.NewInt(0, false)
	}
	return null.NewInt(i, true)
}
