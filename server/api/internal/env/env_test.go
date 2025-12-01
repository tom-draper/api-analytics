package env

import (
	"os"
	"testing"
)

func TestGetIntegerEnvVariable(t *testing.T) {
	t.Run("it returns the default value when the environment variable is not set", func(t *testing.T) {
		defaultValue := 123
		value := GetIntegerEnvVariable("TEST_ENV_VAR", defaultValue)

		if value != defaultValue {
			t.Errorf("Expected %d, got %d", defaultValue, value)
		}
	})

	t.Run("it returns the environment variable value when it is set", func(t *testing.T) {
		expectedValue := 456
		os.Setenv("TEST_ENV_VAR", "456")

		value := GetIntegerEnvVariable("TEST_ENV_VAR", 123)

		if value != expectedValue {
			t.Errorf("Expected %d, got %d", expectedValue, value)
		}

		os.Unsetenv("TEST_ENV_VAR")
	})
}

func TestGetEnvVariable(t *testing.T) {
	t.Run("it returns the default value when the environment variable is not set", func(t *testing.T) {
		defaultValue := "default"
		value := GetEnvVariable("TEST_ENV_VAR", defaultValue)

		if value != defaultValue {
			t.Errorf("Expected %s, got %s", defaultValue, value)
		}
	})

	t.Run("it returns the environment variable value when it is set", func(t *testing.T) {
		expectedValue := "expected"
		os.Setenv("TEST_ENV_VAR", expectedValue)

		value := GetEnvVariable("TEST_ENV_VAR", "default")

		if value != expectedValue {
			t.Errorf("Expected %s, got %s", expectedValue, value)
		}

		os.Unsetenv("TEST_ENV_VAR")
	})
}
