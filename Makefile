test:
	go test -v -coverprofile=cover.out ./...
	go tool cover -func=cover.out
	go tool cover -html=./cover.out -o ./system.html

