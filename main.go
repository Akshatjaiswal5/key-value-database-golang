package main

import (
	"key-value-db-golang/command"
	"key-value-db-golang/datastore"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	db := datastore.NewDatastore()
	cmdHandler := command.NewCommandHandler(db)
	router := gin.Default()

	router.POST("/", cmdHandler.HandleCommand)
	router.Run(":" + os.Getenv("PORT"))
}
