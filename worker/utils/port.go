package utils

import (
	"strconv"
	"strings"
)

func GetPort(addr string) (int, error) {
	return strconv.Atoi(addr[strings.LastIndex(addr, ":")+1:])
}
