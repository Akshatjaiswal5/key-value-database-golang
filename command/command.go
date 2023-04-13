package command

import (
	"key-value-db-golang/datastore"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// request represents the JSON request sent by the client.
type request struct {
	Command string `json:"command"`
}

// commandHandler handles incoming commands.
type commandHandler struct {
	db *datastore.Datastore
}

// NewCommandHandler creates a new command handler with the given datastore.
func NewCommandHandler(db *datastore.Datastore) *commandHandler {
	return &commandHandler{db: db}
}

// HandleCommand receives and parses incoming commands.
func (h *commandHandler) HandleCommand(c *gin.Context) {
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Split the command into parts based on spaces.
	parts := strings.Split(req.Command, " ")
	if len(parts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}

	// Determine which command was sent and call the corresponding method.
	switch parts[0] {
	case "SET":
		h.Set(parts, c)
	case "GET":
		h.Get(parts, c)
	case "QPUSH":
		h.QPush(parts, c)
	case "QPOP":
		h.QPop(parts, c)
	case "BQPOP":
		h.BQPop(parts, c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
	}
}

// Set handles the SET command.
func (h *commandHandler) Set(parts []string, c *gin.Context) {

	if len(parts) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}

	// Parse the key and value from the command.
	key := parts[1]
	value := parts[2]

	// Parse any optional arguments (expiry time and condition).
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

	// Perform the appropriate operation based on the condition.
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
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inserted"})
}

func (h *commandHandler) Get(parts []string, c *gin.Context) {
	// Ensure that the command is properly formatted
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}

	// Extract the key from the command
	key := parts[1]

	// Retrieve the timed value from the datastore
	tValue, err := h.db.Get(key)
	if err != nil {
		// Return an error response if the key is not found in the datastore
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the value associated with the key in the datastore
	c.JSON(http.StatusOK, gin.H{"value": tValue.Value})
}

// QPush adds one or more values to a queue with the given key
func (h *commandHandler) QPush(parts []string, c *gin.Context) {
	// Ensure the correct number of arguments are provided
	if len(parts) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}
	// Get the key and values to be pushed to the queue
	key := parts[1]
	values := parts[2:]

	// Call the QPush method on the datastore to push the values to the queue
	err := h.db.QPush(key, values...)
	if err != nil {
		// Return an error response if there was an error while pushing the values
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Pushed"})
}

// QPop retrieves and removes the top value from the queue with the given key
func (h *commandHandler) QPop(parts []string, c *gin.Context) {
	// Ensure the correct number of arguments are provided
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}
	// Get the key of the queue to pop from
	key := parts[1]

	// Call the QPop method on the datastore to retrieve and remove the top value from the queue
	value, err := h.db.QPop(key)
	if err != nil {
		// Return an error response if the key is not found in the datastore or the queue is empty
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the value associated with the key in the datastore
	c.JSON(http.StatusOK, gin.H{"value": value})
}

// BQPop removes and returns an item from the back of a blocking queue with the given key.
// If the queue is empty, it will block until an item is available or the specified timeout has elapsed.
// If the timeout is 0, it will block indefinitely.
func (h *commandHandler) BQPop(parts []string, c *gin.Context) {
	// Check if the command is valid
	if len(parts) != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid command"})
		return
	}

	// Parse the timeout value from the command
	timeout, err := strconv.Atoi(parts[2])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timeout"})
		return
	}

	// Get the key from the command
	key := parts[1]

	// Remove and return an item from the blocking queue
	value, err := h.db.BQPop(key, time.Duration(timeout)*time.Second)
	if err != nil {
		// Return an error response if the queue is empty or the key is not found in the datastore
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the value associated with the key in the datastore
	c.JSON(http.StatusOK, gin.H{"value": value})
}
