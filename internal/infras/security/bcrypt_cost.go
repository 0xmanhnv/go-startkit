package security

import (
	"os"
	"strconv"
)

func getBcryptCostFromEnv() int {
	v := os.Getenv("BCRYPT_COST")
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 4 || n > 31 {
		return 0
	}
	return n
}
