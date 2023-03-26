package main

import (
	"fmt"

	"server/lib"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := lib.OpenDBConnection()

	app := gin.New()

	r := app.Group("/api")
	r.Use(cors.Default())
	lib.RegisterRouter(r, db) // Register route

	fmt.Println("https://localhost:8080")
	app.Run(":8080")
}
