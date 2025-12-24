build:
	go build -o gosqlit

debug:
	go build -o gosqlit-debug .
	./gosqlit-debug -debug -log debug.log

debug-stdout:
	go build -o gosqlit-debug .
	./gosqlit-debug -debug

debug-run:
	go run main.go -debug -log debug.log

air:
	go install github.com/cosmtrek/air@latest
	air -c .air.toml

run:
	go run main.go

clean:
	rm -f gosqlit gosqlit-debug debug.log

clean-run:
	rm -rf ~/.gosqlit
	make build
	./gosqlit

test:
	go test -v ./...
