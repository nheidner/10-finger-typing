package utils

import (
	"strconv"
	"time"
)

func StringToTime(data string) (time.Time, error) {
	intVal, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(intVal/1000, (intVal%1000)*1e6), nil
}
