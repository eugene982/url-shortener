BUILD_VERSION="v1.20.0"
BUILD_DATE=$(shell date +"%Y/%m/%d %H:%M")
BUILD_COMMIT="Инкремент 20"

BIN_PATH=./bin/shortener
SRC_PATH=./internal
VET_TOOL=./bin/statictest

TEST_BIN=./bin/shortenertestbeta
T_BIN_FLAG=-binary-path=$(BIN_PATH)
T_SRC_FLAG=-source-path=$(SRC_PATH)

tests:
	go test ./...

vet:
	go vet -vettool=$(which $(VET_TOOL)) ./... 

staticcheck:
	$(HOME)/go/bin/staticcheck ./...

buildlint:
	go build -o=bin/staticlint cmd/staticlint/main.go 

lint: buildlint
	bin/staticlint ./...
	
runver:
	go run -ldflags \
		"-X main.buildVersion=$(BUILD_VERSION) -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit="$(BUILD_COMMIT)"' "\
		cmd/shortener/*.go

build: tests vet
	go build -o $(BIN_PATH) \
		-ldflags \
		"-X main.buildVersion=$(BUILD_VERSION) -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit="$(BUILD_COMMIT)"' "\
		cmd/shortener/*.go

#tests
test1:
	$(TEST_BIN) -test.v -test.run=^TestIteration1$$ $(T_BIN_FLAG)

test2:
	$(TEST_BIN) -test.v -test.run=^TestIteration2$$ $(T_SRC_FLAG)
