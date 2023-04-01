package main

import (
	"server/lib"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := lib.OpenDBConnection()

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")
	r.Use(cors.Default())
	lib.RegisterRouter(r, db) // Register route

	app.Run(":3000")
}
