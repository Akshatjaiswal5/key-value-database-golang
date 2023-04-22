# key-value-database-golang
#### This is a simple key-value database implemented in Golang. It supports basic operations such as SET, GET, QPUSH, QPOP and BQPOP. The database uses the Gin web framework to handle incoming requests.

## [Changes from previous submission]
1. Changed how the datastore handled 'queues' It now simply uses channels as queue. The flow of the application is:
QPUSH -> Insert values in channel, QPOP/BQPOP -> Popping values from channel
This simplified the code **however** This made the queue to 'First in First Out' as opposed to 'Last in First Out' required in assignment document.

2. Refactored Handler: Instead of dividing code into different functions for SET,GET and other functions,Now there are 2 functions: A parser function that validates the input and generate a ParsedRequest object with all arguments, and A handler function which recieves the request and sends the response. This achieves separation of concerns.

3. Added Tests: Unit tests for datastore and handler packages.


## Usage

 1. Install Golang on your system. 
 2. Clone the repository and navigate to the root directory. 
 3. Run go build to compile the binary. 
 4. Run  ./key-value-db-golang to start the server. 
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
