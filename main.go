package main

import (
	"key-value-db-golang/datastore"
	"key-value-db-golang/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)


func Handler(w http.ResponseWriter , r *http.Request){
	db := datastore.NewDatastore()

	// Create a new command handler and pass in the datastore
	cmdHandler := handler.NewCommandHandler(db)

	// Create a new Gin router
	router := gin.Default()

	// Map the command handler to the root path for incoming POST requests
	router.POST("/", cmdHandler.HandleCommand)

	// Start the web server and listen for incoming HTTP requests on port 8080
	router.ServeHTTP(w,r)
}
