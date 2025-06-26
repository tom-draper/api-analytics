package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tom-draper/api-analytics/server/api/lib/log"
)

func GetIntegerEnvVariable(name string, defaultValue int) int {
	valueStr := GetEnvVariable(name, strconv.Itoa(defaultValue))

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func GetEnvVariable(name string, defaultValue string) string {
	valueStr := os.Getenv(name)

	if valueStr == "" {
		log.LogToFile(fmt.Sprintf("%s environment variable is blank. Using default value %s=%s.", name, name, defaultValue))
		return defaultValue
	}

	return valueStr
}
