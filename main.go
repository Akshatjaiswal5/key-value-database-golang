package main

import (
	"key-value-db-golang/command"
	"key-value-db-golang/datastore"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new instance of the key-value datastore
	db := datastore.NewDatastore()

	// Create a new command handler and pass in the datastore
	cmdHandler := command.NewCommandHandler(db)

	// Create a new Gin router
	router := gin.Default()

	// Map the command handler to the root path for incoming POST requests
	router.POST("/", cmdHandler.HandleCommand)

	// Start the web server and listen for incoming HTTP requests on port 8080
	router.Run(":8080")
}
