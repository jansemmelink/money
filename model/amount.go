package model

import (
	"strconv"

	"github.com/go-msvc/errors"
)

type Amount float64

func (a *Amount) Parse(s string) error {
	f64, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return errors.Wrapf(err, "invalid amount \"%s\"", s)
	}
	*a = Amount(f64)
	return nil
}
