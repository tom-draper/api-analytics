package usage

import (
	"testing"
)

func TestDatabaseSize(t *testing.T) {
	size, err := DatabaseSize()
	if err != nil {
		t.Error(err)
	}
	if size == "" {
		t.Error("database size is blank")
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
