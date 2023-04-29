package main

import (
	"fmt"
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
)

func main() {
	size, err := usage.DatabaseSize()
	if err != nil {
		panic(err)
	}
	fmt.Println(size)
}
