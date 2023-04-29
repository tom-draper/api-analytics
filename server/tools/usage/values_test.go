package main

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
