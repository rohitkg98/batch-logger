package utils

import (
	"os"
	"strconv"
)

func GetEnvAsInt(key string) int {
	env, exists := os.LookupEnv(key)

	if !exists {
		// panic(fmt.Sprintf("Please set ENV variable: %s", key))
		env = "25"
	}

	envInt, err := strconv.Atoi(env)

	if err != nil {
		panic(err)
	}

	return envInt
}
