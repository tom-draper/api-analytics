package database

import (
	"context"
	"os"
	"testing"
)

var testDB *DB

func TestMain(m *testing.M) {
	// Setup: Create database connection pool once for all tests
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		panic("POSTGRES_URL environment variable is not set")
	}

	var err error
	testDB, err = New(context.Background(), dbURL)
	if err != nil {
		panic("Failed to create database connection: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Teardown: Close database connection
	testDB.Close()

	os.Exit(code)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		apiKey, err := testDB.CreateUser(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if apiKey == "" {
			t.Fatal("Expected non-empty API key")
		}

		// Verify the API key format (should be a UUID)
		if len(apiKey) != 36 {
			t.Errorf("Expected API key length 36, got %d", len(apiKey))
		}

		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	})
}

func TestGetUserID(t *testing.T) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	}()

	t.Run("successful retrieval", func(t *testing.T) {
		userID, err := testDB.GetUserID(ctx, apiKey)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if userID == "" {
			t.Fatal("Expected non-empty user ID")
		}
		if len(userID) != 36 {
			t.Errorf("Expected user ID length 36, got %d", len(userID))
		}
	})

	t.Run("non-existent api key", func(t *testing.T) {
		userID, err := testDB.GetUserID(ctx, "non-existent-key")
		if err == nil {
			t.Fatal("Expected error for non-existent API key")
		}
		if userID != "" {
			t.Errorf("Expected empty user ID, got %s", userID)
		}
	})

	t.Run("empty api key", func(t *testing.T) {
		userID, err := testDB.GetUserID(ctx, "")
		if err == nil {
			t.Fatal("Expected error for empty API key")
		}
		if userID != "" {
			t.Errorf("Expected empty user ID, got %s", userID)
		}
	})
}

func TestGetAPIKey(t *testing.T) {
	ctx := context.Background()

	originalAPIKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if err := testDB.DeleteUser(ctx, originalAPIKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	}()

	userID, err := testDB.GetUserID(ctx, originalAPIKey)
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	t.Run("successful retrieval", func(t *testing.T) {
		apiKey, err := testDB.GetAPIKey(ctx, userID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if apiKey == "" {
			t.Fatal("Expected non-empty API key")
		}
		if apiKey != originalAPIKey {
			t.Errorf("Expected API key %s, got %s", originalAPIKey, apiKey)
		}
	})

	t.Run("non-existent user id", func(t *testing.T) {
		apiKey, err := testDB.GetAPIKey(ctx, "non-existent-user-id")
		if err == nil {
			t.Fatal("Expected error for non-existent user ID")
		}
		if apiKey != "" {
			t.Errorf("Expected empty API key, got %s", apiKey)
		}
	})

	t.Run("empty user id", func(t *testing.T) {
		apiKey, err := testDB.GetAPIKey(ctx, "")
		if err == nil {
			t.Fatal("Expected error for empty user ID")
		}
		if apiKey != "" {
			t.Errorf("Expected empty API key, got %s", apiKey)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		apiKey, err := testDB.CreateUser(ctx)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		userID, err := testDB.GetUserID(ctx, apiKey)
		if err != nil {
			t.Fatalf("Failed to get user ID: %v", err)
		}
		if userID == "" {
			t.Fatal("Expected non-empty user ID")
		}

		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		if _, err := testDB.GetUserID(ctx, apiKey); err == nil {
			t.Fatal("Expected error when getting deleted user")
		}
	})

	t.Run("delete non-existent user", func(t *testing.T) {
		err := testDB.DeleteUser(ctx, "non-existent-key")
		// This should not return an error as DELETE operations are idempotent
		if err != nil {
			t.Errorf("Expected no error for deleting non-existent user, got %v", err)
		}
	})

	t.Run("delete with empty api key", func(t *testing.T) {
		err := testDB.DeleteUser(ctx, "")
		if err != nil {
			t.Errorf("Expected no error for empty API key, got %v", err)
		}
	})
}

func TestDeleteRequests(t *testing.T) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	}()

	t.Run("successful deletion", func(t *testing.T) {
		if err := testDB.DeleteRequests(ctx, apiKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("delete with non-existent api key", func(t *testing.T) {
		if err := testDB.DeleteRequests(ctx, "non-existent-key"); err != nil {
			t.Errorf("Expected no error for non-existent API key, got %v", err)
		}
	})

	t.Run("delete with empty api key", func(t *testing.T) {
		if err := testDB.DeleteRequests(ctx, ""); err != nil {
			t.Errorf("Expected no error for empty API key, got %v", err)
		}
	})
}

func TestDeleteMonitors(t *testing.T) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	}()

	t.Run("successful deletion", func(t *testing.T) {
		if err := testDB.DeleteMonitors(ctx, apiKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("delete with non-existent api key", func(t *testing.T) {
		if err := testDB.DeleteMonitors(ctx, "non-existent-key"); err != nil {
			t.Errorf("Expected no error for non-existent API key, got %v", err)
		}
	})

	t.Run("delete with empty api key", func(t *testing.T) {
		if err := testDB.DeleteMonitors(ctx, ""); err != nil {
			t.Errorf("Expected no error for empty API key, got %v", err)
		}
	})
}

func TestDeletePings(t *testing.T) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	}()

	t.Run("successful deletion", func(t *testing.T) {
		if err := testDB.DeletePings(ctx, apiKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("delete with non-existent api key", func(t *testing.T) {
		if err := testDB.DeletePings(ctx, "non-existent-key"); err != nil {
			t.Errorf("Expected no error for non-existent API key, got %v", err)
		}
	})

	t.Run("delete with empty api key", func(t *testing.T) {
		if err := testDB.DeletePings(ctx, ""); err != nil {
			t.Errorf("Expected no error for empty API key, got %v", err)
		}
	})
}

func TestCheckConnection(t *testing.T) {
	ctx := context.Background()

	t.Run("successful connection check", func(t *testing.T) {
		if err := testDB.CheckConnection(ctx); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestUserWorkflow(t *testing.T) {
	ctx := context.Background()

	t.Run("complete user lifecycle", func(t *testing.T) {
		apiKey, err := testDB.CreateUser(ctx)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
		if apiKey == "" {
			t.Fatal("Expected non-empty API key")
		}

		userID, err := testDB.GetUserID(ctx, apiKey)
		if err != nil {
			t.Fatalf("Failed to get user ID: %v", err)
		}
		if userID == "" {
			t.Fatal("Expected non-empty user ID")
		}

		retrievedAPIKey, err := testDB.GetAPIKey(ctx, userID)
		if err != nil {
			t.Fatalf("Failed to get API key: %v", err)
		}
		if retrievedAPIKey != apiKey {
			t.Errorf("Expected API key %s, got %s", apiKey, retrievedAPIKey)
		}

		// Delete all associated data
		if err := testDB.DeleteRequests(ctx, apiKey); err != nil {
			t.Errorf("Failed to delete requests: %v", err)
		}

		if err := testDB.DeleteMonitors(ctx, apiKey); err != nil {
			t.Errorf("Failed to delete monitors: %v", err)
		}

		if err := testDB.DeletePings(ctx, apiKey); err != nil {
			t.Errorf("Failed to delete pings: %v", err)
		}

		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to delete user: %v", err)
		}

		// Verify user is gone
		if _, err := testDB.GetUserID(ctx, apiKey); err == nil {
			t.Fatal("Expected error when getting deleted user")
		}
	})
}

func TestDeleteAllData(t *testing.T) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer func() {
		if err := testDB.DeleteUser(ctx, apiKey); err != nil {
			t.Errorf("Failed to clean up test user: %v", err)
		}
	}()

	t.Run("successful deletion of all data", func(t *testing.T) {
		if err := testDB.DeleteAllData(ctx, apiKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestDeleteUserAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion of user account", func(t *testing.T) {
		apiKey, err := testDB.CreateUser(ctx)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		if err := testDB.DeleteUserAccount(ctx, apiKey); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify user is gone
		if _, err := testDB.GetUserID(ctx, apiKey); err == nil {
			t.Fatal("Expected error when getting deleted user")
		}
	})
}

func isValidUUID(uuid string) bool {
	if len(uuid) != 36 {
		return false
	}

	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		return false
	}

	// Check that other characters are hexadecimal
	for i, char := range uuid {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue // Skip hyphens
		}
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

func TestUUIDValidation(t *testing.T) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer testDB.DeleteUser(ctx, apiKey)

	if !isValidUUID(apiKey) {
		t.Errorf("Generated API key %s is not a valid UUID", apiKey)
	}

	invalidUUIDs := []string{
		"",
		"too-short",
		"12345678-1234-1234-1234-12345678901",   // too short
		"12345678-1234-1234-1234-1234567890123", // too long
		"12345678_1234_1234_1234_123456789012",  // wrong separators
		"1234567g-1234-1234-1234-123456789012",  // invalid hex character
	}

	for _, invalid := range invalidUUIDs {
		if isValidUUID(invalid) {
			t.Errorf("UUID %s should be invalid", invalid)
		}
	}
}

func BenchmarkCreateUser(b *testing.B) {
	ctx := context.Background()
	var apiKeys []string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		apiKey, err := testDB.CreateUser(ctx)
		if err != nil {
			b.Fatal(err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	b.StopTimer()
	for _, apiKey := range apiKeys {
		testDB.DeleteUser(ctx, apiKey)
	}
}

func BenchmarkGetUserID(b *testing.B) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer testDB.DeleteUser(ctx, apiKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.GetUserID(ctx, apiKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetAPIKey(b *testing.B) {
	ctx := context.Background()

	apiKey, err := testDB.CreateUser(ctx)
	if err != nil {
		b.Fatal(err)
	}
	defer testDB.DeleteUser(ctx, apiKey)

	userID, err := testDB.GetUserID(ctx, apiKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := testDB.GetAPIKey(ctx, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}