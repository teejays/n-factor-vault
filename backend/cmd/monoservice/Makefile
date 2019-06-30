PROJECT_NAME=monoservice
BIN_NAME=$(PROJECT_NAME).bin

GO=go
GO_BUILD=$(GO) build

all: build run

build:
	$(GO_BUILD) -o $(BIN_NAME) main.go 

run: 
	./$(BIN_NAME)