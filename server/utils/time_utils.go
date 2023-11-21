package utils

import (
	"10-typing/errors"
	"strconv"
	"time"
)

func StringToTime(data string) (time.Time, error) {
	const op errors.Op = "utils.StringToTime"

	intVal, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return time.Time{}, errors.E(op, err)
	}

	return time.Unix(intVal/1000, (intVal%1000)*1e6), nil
}
