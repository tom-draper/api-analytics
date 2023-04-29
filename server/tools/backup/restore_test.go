package main

import (
	"testing"
)

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			if !found {
				diff = append(diff, s1)
			}
		}
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func directories() []string {
	entries, err := os.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	var dirs []string
	for _, e := range entries {
		dirs = append(dirs, e.Name())
	}
	return dirs
}

func TestRestore(t *testing.T) {
	dirsBefore := directories()
	BackupDatabase()
	dirsAfter := directories()
	diff := difference(dirsBefore, dirsAfter)
	backupDir = diff[0]
	Restore(backupDir, "test")
}
