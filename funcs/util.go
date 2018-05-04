package funcs

import (
	"fmt"
	"strconv"
)

func GetInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			fval, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return 0
			}
			return int(fval)
		}
		return int(n)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func GetString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}
