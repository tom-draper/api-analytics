package main

import (
	"github.com/tom-draper/api-analytics/server/api/routes"
	"github.com/tom-draper/api-analytics/server/database"

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
