package main

import (
	"banking-app/configs"
	"banking-app/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	configs.ConnectDB()
	routes.UserRoute(r)
	r.Run()
}
