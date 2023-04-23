package main

import (
	"key-value-db-golang/datastore"
	"key-value-db-golang/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

var(
	app *gin.Engine
)

func myRoute(r *gin.RouterGroup){
 db := datastore.NewDatastore()
	cmdHandler := handler.NewCommandHandler(db)
	r.POST("/",cmdHandler.HandleCommand)
}

func init(){
	app = gin.New()
	r := app.Group("/")
	myRoute(r)

}

// ADD THIS SCRIPT
func Handler(w http.ResponseWriter , r *http.Request){
	app.ServeHTTP(w,r)
}
