ensure::
	go mod tidy -go=1.16 && go mod tidy -go=1.18

lint::
	golangci-lint run -c .golangci.yml --timeout 10m