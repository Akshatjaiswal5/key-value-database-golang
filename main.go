package main

import (
	"key-value-db-golang/command"
	"key-value-db-golang/datastore"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {
	db := datastore.NewDatastore()
	cmdHandler := command.NewCommandHandler(db)
	router = gin.Default()

	router.POST("/", cmdHandler.Handler)
	router.Run(":" + os.Getenv("PORT"))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)

}
