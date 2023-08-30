staticcheck:
	$(HOME)/go/bin/staticcheck ./...

buildstaticlint:
	go build -o=bin/staticlint cmd/staticlint/main.go 

staticlint: buildstaticlint
	bin/staticlint ./...