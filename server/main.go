package main

import (
	"server/lib/database"
	"server/lib/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := database.OpenDBConnection()

	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	r := app.Group("/api")
	r.Use(cors.Default())
	routes.RegisterRouter(r, db)

	app.Run(":3000")
}
