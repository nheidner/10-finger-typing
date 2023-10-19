package utils

func SliceContains[T string | int | bool](s []string, item string) bool {
	for _, a := range s {
		if a == item {
			return true
		}
	}

	return false
}
