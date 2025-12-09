DEFAULT_GOAL: run

tidy:
	go mod tidy

run: tidy
	go run main.go

build: tidy
	go build -o main main.go
