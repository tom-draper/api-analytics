package main

import (
	"fmt"
	"github.com/tom-draper/api-analytics/server/tools/usage/usage"
)

func main() {
	monitors := usage.Monitors()
	fmt.Println(monitors)
}
