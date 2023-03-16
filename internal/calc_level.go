package internal

import (
	"errors"
	"strconv"
)

func calc2NPow(v int64) int64 {
	for i := 0; i < 32; i++ {
		if v>>i < 2 {
			return int64(i)
		}
	}
	return 32
}

func calcLevel(mortgaged string) (int64, error) {
	valueLen := len(mortgaged)
	if valueLen < 18 {
		return 0, errors.New("invalid fmt")
	}
	mortgageIntStr := mortgaged[:valueLen-18]
	int64Mortgaged, err := strconv.ParseInt(mortgageIntStr, 10, 64)
	if err != nil {
		return 0, err
	}
	hundredInt := int64Mortgaged / 100
	return calc2NPow(hundredInt), nil
}

func CalcLimit(mortgaged string) (int64, error) {
	level, err := calcLevel(mortgaged)
	if err != nil {
		return 100, err
	}
	limit := 1000 / ((level + 1) * 10)

	return limit, nil
}

func CalcLimitByLevel(level int64) (int64, error) {
	if level == 0 {
		return 1000, nil
	}
	limit := 1000 / ((level + 1) * 10)

	return limit, nil
}
