TESTPARALLELISM := 4

ensure::
	go mod tidy && go mod download

lint::
	golangci-lint run -c .golangci.yml --timeout 10m

test::
	cd pkg && go test -v -count=1 -cover -timeout 2h -parallel ${TESTPARALLELISM} ./...
