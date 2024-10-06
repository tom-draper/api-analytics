package usage

import (
	"context"
	"testing"
)

func TestTableSize(t *testing.T) {
	size, err := TableSize(context.Background(), "requests")
	if err != nil {
		t.Error(err)
	}
	if size == "" {
		t.Error("database size is blank")
	}
}

func TestTableColumnSize(t *testing.T) {
	totalSize, tableSize, percentage, err := TableColumnSize(context.Background(), "requests", "api_key")
	if err != nil {
		t.Error(err)
	}
	if totalSize == "" || tableSize == "" {
		t.Error("database column size is blank")
	}
	if percentage == 0 {
		t.Error("database column size percentage is 0")
	}
}

func TestDatabaseConnections(t *testing.T) {
	connections, err := DatabaseConnections(context.Background())
	if err != nil {
		t.Error(connections)
	}
	if connections == 0 {
		t.Error("number of active database connections is 0")
	}
}
