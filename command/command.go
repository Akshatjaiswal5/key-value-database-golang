package command

import (
	"key-value-db-golang/datastore"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type request struct {
	Command string `json:"command"`
}

type commandHandler struct {
	db *datastore.Datastore
}

func NewCommandHandler(db *datastore.Datastore) *commandHandler {
	return &commandHandler{db: db}
}

func (h *commandHandler) HandleCommand(c *gin.Context) {
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parts := strings.Split(req.Command, " ")
	if len(parts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}
	switch parts[0] {
	case "SET":
		h.Set(parts, c)
	case "GET":
		h.Get(parts, c)
	case "QPUSH":
		h.Set(parts, c)
	case "QPOP":
		h.Get(parts, c)
	case "BQPOP":
		h.Set(parts, c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
	}
}

func (h *commandHandler) Set(parts []string, c *gin.Context) {

	if len(parts) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}
	key := parts[1]
	value := parts[2]
	expirySeconds := 0
	condition := ""
	for i := 3; i < len(parts); i++ {
		if strings.HasPrefix(parts[i], "EX") {
			if expiry, err := strconv.Atoi(parts[i][2:]); err == nil {
				expirySeconds = expiry
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
				return
			}
		} else if parts[i] == "NX" || parts[i] == "XX" {
			condition = parts[i]
		}
	}

	if condition == "" {
		h.db.Set(key, value, expirySeconds)
	} else {
		_, err := h.db.Get(key)
		if err == nil && condition == "NX" {
			h.db.Set(key, value, expirySeconds)
		} else if err != nil && condition == "XX" {
			h.db.Set(key, value, expirySeconds)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Not Inserted"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inserted"})
}
func (h *commandHandler) Get(parts []string, c *gin.Context) {

	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}

	key := parts[1]

	tValue, err := h.db.Get(key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": tValue.Value})

}
func (h *commandHandler) QPush(parts []string, c *gin.Context) {
}
func (h *commandHandler) QPop(parts []string, c *gin.Context) {
}
func (h *commandHandler) BQPop(parts []string, c *gin.Context) {
}
