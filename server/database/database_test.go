package database

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment state
	originalURL := os.Getenv("POSTGRES_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("POSTGRES_URL", originalURL)
		} else {
			os.Unsetenv("POSTGRES_URL")
		}
		// Reset package variable
		dbURL = ""
	}()

	t.Run("loads config successfully with POSTGRES_URL set", func(t *testing.T) {
		// Reset package variable
		dbURL = ""

		testURL := "postgres://user:pass@localhost:5432/testdb"
		os.Setenv("POSTGRES_URL", testURL)

		err := LoadConfig()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if dbURL != testURL {
			t.Errorf("expected dbURL to be %s, got %s", testURL, dbURL)
		}
	})

	t.Run("returns error when POSTGRES_URL is not set", func(t *testing.T) {
		// Reset package variable
		dbURL = ""

		os.Unsetenv("POSTGRES_URL")

		err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when POSTGRES_URL is not set")
		}

		expectedError := "POSTGRES_URL is not set in the environment"
		if err.Error() != expectedError {
			t.Errorf("expected error message %q, got %q", expectedError, err.Error())
		}
	})

	t.Run("returns error when POSTGRES_URL is empty string", func(t *testing.T) {
		// Reset package variable
		dbURL = ""

		os.Setenv("POSTGRES_URL", "")

		err := LoadConfig()
		if err == nil {
			t.Fatal("expected error when POSTGRES_URL is empty")
		}

		expectedError := "POSTGRES_URL is not set in the environment"
		if err.Error() != expectedError {
			t.Errorf("expected error message %q, got %q", expectedError, err.Error())
		}
	})
}

func TestNewConnection(t *testing.T) {
	// Save original environment state
	originalURL := os.Getenv("POSTGRES_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("POSTGRES_URL", originalURL)
		} else {
			os.Unsetenv("POSTGRES_URL")
		}
		// Reset package variable
		dbURL = ""
	}()

	t.Run("returns error when POSTGRES_URL not configured", func(t *testing.T) {
		// Reset package variable
		dbURL = ""

		os.Unsetenv("POSTGRES_URL")

		conn, err := NewConnection()
		if err == nil {
			t.Fatal("expected error when POSTGRES_URL is not set")
		}
		if conn != nil {
			t.Error("expected connection to be nil when error occurs")
		}

		expectedError := "POSTGRES_URL is not set in the environment"
		if err.Error() != expectedError {
			t.Errorf("expected error message %q, got %q", expectedError, err.Error())
		}
	})

	t.Run("attempts connection with valid URL format", func(t *testing.T) {
		// Reset package variable
		dbURL = ""

		// Use a URL that has correct format but points to non-existent server
		// This will test that the function attempts to connect but fails gracefully
		testURL := "postgres://user:pass@nonexistent:5432/testdb"
		os.Setenv("POSTGRES_URL", testURL)

		conn, err := NewConnection()
		if err == nil {
			// If somehow it connects (shouldn't happen with nonexistent host), close it
			if conn != nil {
				conn.Close(nil)
			}
			t.Skip("connection unexpectedly succeeded - skipping this test case")
		}
		if conn != nil {
			t.Error("expected connection to be nil when connection fails")
		}

		// The exact error message will depend on the system, but it should contain
		// some indication of connection failure
		if err.Error() == "POSTGRES_URL is not set in the environment" {
			t.Error("got config error instead of connection error")
		}
	})

	t.Run("uses existing dbURL when already set", func(t *testing.T) {
		// Pre-set the package variable
		dbURL = "postgres://user:pass@presethost:5432/testdb"

		// Set a different environment variable to ensure it uses the preset one
		os.Setenv("POSTGRES_URL", "postgres://user:pass@envhost:5432/testdb")

		conn, err := NewConnection()
		// Should attempt to connect to presethost, not envhost
		// Since presethost doesn't exist, it should fail
		if err == nil {
			if conn != nil {
				conn.Close(nil)
			}
			t.Skip("connection unexpectedly succeeded - skipping this test case")
		}
		if conn != nil {
			t.Error("expected connection to be nil when connection fails")
		}

		// Should not be a config error since dbURL was already set
		if err.Error() == "POSTGRES_URL is not set in the environment" {
			t.Error("got config error when dbURL was already set")
		}
	})

	t.Run("returns error with invalid URL format", func(t *testing.T) {
		// Reset package variable
		dbURL = ""

		invalidURL := "not-a-valid-url"
		os.Setenv("POSTGRES_URL", invalidURL)

		conn, err := NewConnection()
		if err == nil {
			if conn != nil {
				conn.Close(nil)
			}
			t.Fatal("expected error with invalid URL format")
		}
		if conn != nil {
			t.Error("expected connection to be nil when connection fails")
		}
	})
}

func BenchmarkLoadConfig(b *testing.B) {
	originalURL := os.Getenv("POSTGRES_URL")
	os.Setenv("POSTGRES_URL", "postgres://user:pass@localhost:5432/testdb")
	defer func() {
		if originalURL != "" {
			os.Setenv("POSTGRES_URL", originalURL)
		} else {
			os.Unsetenv("POSTGRES_URL")
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dbURL = "" // Reset for each iteration
		LoadConfig()
	}
}
