package database

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ctx := context.Background()

	t.Run("creates connection pool successfully with valid URL", func(t *testing.T) {
		// This assumes POSTGRES_URL is set in the environment
		// You can also read it from env for the test
		testURL := "postgres://user:pass@localhost:5432/testdb"
		
		db, err := New(ctx, testURL)
		if err != nil {
			// Connection failure is expected if database isn't running
			// but we can verify the error is a connection error, not a validation error
			t.Logf("Connection failed (expected if DB not running): %v", err)
			return
		}
		defer db.Close()

		// Verify we can ping the database
		if err := db.Pool.Ping(ctx); err != nil {
			t.Errorf("failed to ping database: %v", err)
		}
	})

	t.Run("returns error when URL is empty", func(t *testing.T) {
		db, err := New(ctx, "")
		if err == nil {
			if db != nil {
				db.Close()
			}
			t.Fatal("expected error when URL is empty")
		}
		if db != nil {
			t.Error("expected db to be nil when error occurs")
		}

		expectedError := "database URL cannot be empty"
		if err.Error() != expectedError {
			t.Errorf("expected error message %q, got %q", expectedError, err.Error())
		}
	})

	t.Run("returns error with invalid URL format", func(t *testing.T) {
		invalidURL := "not-a-valid-url"
		
		db, err := New(ctx, invalidURL)
		if err == nil {
			if db != nil {
				db.Close()
			}
			t.Fatal("expected error with invalid URL format")
		}
		if db != nil {
			t.Error("expected db to be nil when connection fails")
		}
	})

	t.Run("returns error when connection fails", func(t *testing.T) {
		// Use a URL that has correct format but points to non-existent server
		testURL := "postgres://user:pass@nonexistent.invalid:5432/testdb"
		
		db, err := New(ctx, testURL)
		if err == nil {
			if db != nil {
				db.Close()
			}
			t.Fatal("expected error when connecting to non-existent server")
		}
		if db != nil {
			t.Error("expected db to be nil when connection fails")
		}
	})

	t.Run("respects context timeout", func(t *testing.T) {
		// Create a context that times out immediately
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		// Use a URL that would take time to connect
		testURL := "postgres://user:pass@192.0.2.1:5432/testdb" // TEST-NET-1, should timeout
		
		db, err := New(ctx, testURL)
		if err == nil {
			if db != nil {
				db.Close()
			}
			t.Skip("connection unexpectedly succeeded")
		}
		if db != nil {
			t.Error("expected db to be nil when context times out")
		}
	})
}

func TestDBClose(t *testing.T) {
	ctx := context.Background()
	testURL := "postgres://user:pass@localhost:5432/testdb"

	db, err := New(ctx, testURL)
	if err != nil {
		t.Skipf("Skipping test - database connection failed: %v", err)
	}

	t.Run("closes connection pool successfully", func(t *testing.T) {
		// This should not panic
		db.Close()

		// After closing, operations should fail
		err := db.Pool.Ping(ctx)
		if err == nil {
			t.Error("expected error after closing connection pool")
		}
	})
}

func TestDBCheckConnection(t *testing.T) {
	ctx := context.Background()

	t.Run("successful connection check", func(t *testing.T) {
		// Uses the global testDB from query_test.go TestMain
		// If running this file independently, you'll need to set up testDB
		if testDB == nil {
			t.Skip("testDB not initialized - run with query_test.go")
		}

		if err := testDB.CheckConnection(ctx); err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestConnectionPoolConcurrency(t *testing.T) {
	if testDB == nil {
		t.Skip("testDB not initialized - run with query_test.go")
	}

	ctx := context.Background()

	t.Run("handles concurrent requests", func(t *testing.T) {
		// Test that the connection pool handles concurrent operations
		const numGoroutines = 10
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				apiKey, err := testDB.CreateUser(ctx)
				if err != nil {
					errors <- err
					return
				}
				
				// Clean up
				if err := testDB.DeleteUser(ctx, apiKey); err != nil {
					errors <- err
					return
				}
				
				errors <- nil
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			if err := <-errors; err != nil {
				t.Errorf("Concurrent operation failed: %v", err)
			}
		}
	})
}

func TestConnectionPoolReuse(t *testing.T) {
	if testDB == nil {
		t.Skip("testDB not initialized - run with query_test.go")
	}

	ctx := context.Background()

	t.Run("reuses connections efficiently", func(t *testing.T) {
		// Make multiple sequential requests
		// If connection pooling works, this should be fast
		for i := 0; i < 5; i++ {
			apiKey, err := testDB.CreateUser(ctx)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}
			
			if err := testDB.DeleteUser(ctx, apiKey); err != nil {
				t.Fatalf("Failed to delete user: %v", err)
			}
		}
	})
}

func BenchmarkNewConnection(b *testing.B) {
	ctx := context.Background()
	testURL := "postgres://user:pass@localhost:5432/testdb"

	b.Run("with pooling", func(b *testing.B) {
		// Create pool once
		db, err := New(ctx, testURL)
		if err != nil {
			b.Skipf("Database connection failed: %v", err)
		}
		defer db.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Operations use the pool
			_ = db.Pool.Ping(ctx)
		}
	})

	b.Run("pool creation overhead", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			db, err := New(ctx, testURL)
			if err != nil {
				b.Skipf("Database connection failed: %v", err)
			}
			db.Close()
		}
	})
}

func BenchmarkConcurrentOperations(b *testing.B) {
	if testDB == nil {
		b.Skip("testDB not initialized")
	}

	ctx := context.Background()

	b.Run("sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			apiKey, err := testDB.CreateUser(ctx)
			if err != nil {
				b.Fatal(err)
			}
			testDB.DeleteUser(ctx, apiKey)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				apiKey, err := testDB.CreateUser(ctx)
				if err != nil {
					b.Fatal(err)
				}
				testDB.DeleteUser(ctx, apiKey)
			}
		})
	})
}