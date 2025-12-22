build:
	go build -o gosqlit

air:
	go install github.com/cosmtrek/air@latest
	air -c .air.toml

run:
	go run main.go

clean-run:
	rm -rf ~/.gosqlit
	make build
	./gosqlit

test:
	go test -v ./...
