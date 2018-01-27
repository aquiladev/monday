package util

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"
)

func Parse(value string) (*big.Int, error) {
	num := big.NewInt(0)
	if _, ok := num.SetString(value, 10); !ok {
		return nil, errors.New(fmt.Sprintf("Error parsing line %#v\n", value))
	}
	return num, nil
}
