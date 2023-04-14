package handler

import (
	"key-value-db-golang/command"
	"key-value-db-golang/datastore"
	"net/http"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func Handler(w http.ResponseWriter, r *http.Request) {
	db := datastore.NewDatastore()
	cmdHandler := command.NewCommandHandler(db)
	router = gin.Default()

	router.POST("/", cmdHandler.Handler)

	router.ServeHTTP(w, r)

}
