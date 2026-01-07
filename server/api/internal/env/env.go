package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
)

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Info("Failed to load .env file")
	}
	return err
}

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
		log.Info(fmt.Sprintf("%s environment variable is blank. Using default value %s=%s.", name, name, defaultValue))
		return defaultValue
	}

	return valueStr
}
