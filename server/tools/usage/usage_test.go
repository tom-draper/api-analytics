package main

import (
	"testing"
)

func TestUsage(t *testing.T) {
	requests, err := Usage()
	if err != nil {
		t.Error(err)
	}
}

func TestUsers(t *testing.T) {
	users, err := Users()
	if err != nil {
		t.Error(err)
	}
}

func TestTopUsers(t *testing.T) {
	users, err := TopUsers()
	if err != nil {
		t.Error(err)
	}
}

func TestTotalRequests(t *testing.T) {
	requests, err := TotalRequests()
	if err != nil {
		t.Error(err)
	}
}
