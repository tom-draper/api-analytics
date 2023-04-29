package usage

import (
	"testing"
)

func TestUserAgents(t *testing.T) {
	userAgents, err := UserAgents()
	if err != nil {
		t.Error(err)
	}
}

func TestIPAddresses(t *testing.T) {
	ipAddresses, err := IPAddresses()
	if err != nil {
		t.Error(err)
	}
}

func TestLocations(t *testing.T) {
	locations, err := Locations()
	if err != nil {
		t.Error(err)
	}
}

func TestAvgResponseTime(t *testing.T) {
	avg, err := AvgResponseTime()
	if err != nil {
		t.Error(err)
	}
}
