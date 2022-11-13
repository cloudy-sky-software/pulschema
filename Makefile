ensure::
	go mod tidy && go mod download

lint::
	golangci-lint run -c .golangci.yml --timeout 10m