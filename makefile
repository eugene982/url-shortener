BUILD_VERSION="v1.20.0"
BUILD_DATE=$(shell date +"%Y/%m/%d %H:%M")
BUILD_COMMIT="Инкремент 20"

staticcheck:
	$(HOME)/go/bin/staticcheck ./...

buildstaticlint:
	go build -o=bin/staticlint cmd/staticlint/main.go 

staticlint: buildstaticlint
	bin/staticlint ./...
	
runver:
	go run -ldflags \
		"-X main.buildVersion=$(BUILD_VERSION) -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit="$(BUILD_COMMIT)"' "\
		cmd/shortener/main.go