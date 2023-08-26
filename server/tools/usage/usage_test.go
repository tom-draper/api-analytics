package usage

import (
	"testing"
)

func TestTableSize(t *testing.T) {
	size, err := TableSize("requests")
	if err != nil {
		t.Error(err)
	}
	if size == "" {
		t.Error("database size is blank")
	}
}

func TestTableColumnSize(t *testing.T) {
	size, err := TableColumnSize("requests", "api_key")
	if err != nil {
		t.Error(err)
	}
	if size == "" {
		t.Error("database column size is blank")
	}
}

func TestDatabaseConnections(t *testing.T) {
	connections, err := DatabaseConnections()
	if err != nil {
		t.Error(connections)
	}
	if connections == 0 {
		t.Error("number of active database connections is 0")
	}
}
