# key-value-database-golang
#### This is a simple key-value database implemented in Golang. It supports basic operations such as SET, GET, QPUSH, QPOP and BQPOP. The database uses the Gin web framework to handle incoming requests.

## Usage

 1. Install Golang on your system. 
 2. Clone the repository and navigate to
 3. the root directory. 
 4. Run go build to compile the binary. 
 5. Run  ./key-value-db-golang to start the server. 
                The server is now listening on port 8080.

## Supported Commands

 - SET key value - Sets a key to a given value. 
 - GET key - Returns the value associated with a key. 
 - QPUSH key value [value2 ...] - Pushes one or more values onto a queue. 
 - QPOP key - Pops a value from a queue. 
 - BQPOP key timeout - Blocks the thread until a value is read from the queue. The command fails if multiple clients try to read from the queue at the same time. Timeout is specified in seconds.

## Example Requests

    SET name John

Sets the value of the name key to John.

    GET name

Returns the value associated with the name key, which is John.

    QPUSH queue1 value1 value2

Pushes value1 and value2 onto the queue1 queue.

    QPOP queue1

Pops a value from the queue1 queue.

    BQPOP queue1 5

Blocks the thread until a value is read from the queue1 queue, or until 5 seconds have elapsed.
