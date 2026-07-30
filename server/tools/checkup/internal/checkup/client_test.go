package checkup

import (
	"context"
	"testing"
)

func TestNewClientFromEnvRequiresPostgresURL(t *testing.T) {
	// With POSTGRES_URL unset and no required fields, NewClientFromEnv must return
	// a real error and a nil client — never (nil, nil), which the caller would
	// mistake for success and then dereference.
	t.Setenv("POSTGRES_URL", "")

	client, err := NewClientFromEnv(context.Background())
	if err == nil {
		t.Fatal("expected an error when POSTGRES_URL is unset")
	}
	if client != nil {
		t.Errorf("expected a nil client alongside the error, got %v", client)
	}
}
