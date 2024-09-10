# Variables for binaries
SERVER_BINARY=./server/server
CLIENT_BINARY=./client/client

# Build targets
.PHONY: all build-server build-client clean run

all: build-server build-client

build-server:
	@echo "Building server..."
	go build -o $(SERVER_BINARY) ./server/server.go

build-client:
	@echo "Building client..."
	go build -o $(CLIENT_BINARY) ./client/client.go

# Run command
run: build-server build-client
	@echo "Starting server in the background..."
	$(SERVER_BINARY) & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running client..."; \
	$(CLIENT_BINARY); \
	echo "Client finished. Stopping server and cleaning up..."; \
	kill $$SERVER_PID;

clean:
	@echo "Cleaning up binaries..."
	rm -f $(SERVER_BINARY) $(CLIENT_BINARY)
