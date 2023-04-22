package handler

import (
	"errors"
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

type parsedRequest struct {
	command       string
	key           string
	value         string
	condition     string
	expirySeconds time.Duration
	values        []string
	timeout       time.Duration
}

// commandHandler handles incoming commands.
type commandHandler struct {
	db *datastore.Datastore
}

// NewCommandHandler creates a new command handler with the given datastore.
func NewCommandHandler(db *datastore.Datastore) *commandHandler {
	return &commandHandler{db: db}
}

// parseRequest parses and validate the incoming request
func ParseRequest(c *gin.Context) (*parsedRequest, error) {
	var req request
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// Split the command into parts based on spaces.
	parts := strings.Split(req.Command, " ")
	if len(parts) == 0 {
		return nil, errors.New("Invalid number of arguments")
	}

	// Determine which command was sent and return the appropiate parsedRequest.
	switch parts[0] {
	case "SET":
		if len(parts) < 3 {
			return nil, errors.New("Invalid number of arguments for SET")
		}
		// Parse the key and value.
		key := parts[1]
		value := parts[2]
		expirySeconds := 0
		condition := ""
		// Parse any optional arguments (expiry time and condition).
		for i := 3; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], "EX") {
				if expiry, err := strconv.Atoi(parts[i][2:]); err == nil {
					expirySeconds = expiry
				} else {
					return nil, errors.New("Invalid arguments for SET")
				}
			} else if parts[i] == "NX" || parts[i] == "XX" {
				condition = parts[i]
			}
		}

		if expirySeconds < 0 {
			return nil, errors.New("Invalid expirySeconds")
		}
		return &parsedRequest{
			command:       "SET",
			key:           key,
			value:         value,
			expirySeconds: time.Duration(expirySeconds),
			condition:     condition,
		}, nil

	case "GET":
		if len(parts) != 2 {
			return nil, errors.New("Invalid number of arguments for GET")
		}
		key := parts[1] // Extract the key from the command
		return &parsedRequest{
			command: "GET",
			key:     key,
		}, nil

	case "QPUSH":
		if len(parts) < 3 {
			return nil, errors.New("Invalid number of arguments for QPUSH")
		}
		// Get the key and values to be pushed to the queue
		key := parts[1]
		values := parts[2:]
		return &parsedRequest{
			command: "QPUSH",
			key:     key,
			values:  values,
		}, nil

	case "QPOP":
		if len(parts) != 2 {
			return nil, errors.New("Invalid number of arguments for QPOP")
		}
		key := parts[1] // Extract the key from the command
		return &parsedRequest{
			command: "QPOP",
			key:     key,
		}, nil

	case "BQPOP":
		if len(parts) != 3 {
			return nil, errors.New("Invalid number of arguments for BQPOP")
		}
		key := parts[1]                        // Get the key from the command
		timeout, err := strconv.Atoi(parts[2]) // Parse the timeout value from the command
		if err != nil {
			return nil, errors.New("Invalid value for timeout")
		}
		return &parsedRequest{
			command: "BQPOP",
			key:     key,
			timeout: time.Duration(timeout),
		}, nil

	default:
		return nil, errors.New("Unrecognised command")
	}
}

// HandleCommand receives and responds to incoming commands.
func (h *commandHandler) HandleCommand(c *gin.Context) {
	req, err := ParseRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// Responding to request
	switch req.command {
	case "SET":
		if req.condition == "" {
			h.db.Set(req.key, req.value, req.expirySeconds)
			c.JSON(http.StatusOK, gin.H{"message": "Inserted"})
		} else if _, err := h.db.Get(req.key); (err == nil && req.condition == "NX") || (err != nil && req.condition == "XX") {
			h.db.Set(req.key, req.value, req.expirySeconds)
			c.JSON(http.StatusOK, gin.H{"message": "Inserted"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Not Inserted"})
		}
	case "GET":
		value, err := h.db.Get(req.key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Return an error response if the key is not found in the datastore
		} else {
			c.JSON(http.StatusOK, gin.H{"value": value}) // Return the value associated with the key in the datastore
		}
	case "QPUSH":
		err := h.db.QPush(req.key, req.values...)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Return an error response if there was an error while pushing the values
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "Pushed successfully"}) // Return a success response
		}
	case "QPOP":
		value, err := h.db.QPop(req.key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // // Return an error response if the key is not found in the datastore or the queue is empty
		} else {
			c.JSON(http.StatusOK, gin.H{"value": value}) // Return the value associated with the key in the datastore
		}
	case "BQPOP":
		value, err := h.db.BQPop(req.key, req.timeout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // // Return an error response if the key is not found in the datastore or the queue is empty
		} else {
			c.JSON(http.StatusOK, gin.H{"value": value}) // Return the value associated with the key in the datastore
		}
	default:
		c.JSON(http.StatusOK, gin.H{"value": req.value})
	}

}
