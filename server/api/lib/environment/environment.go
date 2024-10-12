package environment

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/api/lib/log"
)

func GetIntegerEnvVariable(name string, defaultValue int) int {
	err := godotenv.Load(".env")
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to load .env file. Using default value %s=%d.", name, defaultValue))
		return defaultValue
	}

	valueStr := os.Getenv(name)
	if valueStr == "" {
		log.LogToFile(fmt.Sprintf("%s environment variable is blank. Using default value %s=%d.", name, name, defaultValue))
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.LogToFile(fmt.Sprintf("%s environment variable is not an integer. Using default value %s=%d.", name, name, defaultValue))
		return defaultValue
	}

	return value
}

func GetEnvVariable(name string, defaultValue string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.LogToFile(fmt.Sprintf("Failed to load .env file. Using default value %s=%s.", name, defaultValue))
		return defaultValue
	}

	value := os.Getenv(name)
	if value == "" {
		log.LogToFile(fmt.Sprintf("%s environment variable is blank. Using default value %s=%s.", name, name, defaultValue))
		return defaultValue
	}

	return value
}
