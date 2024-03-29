BUILD_VERSION="v1.25.3"
BUILD_DATE=$(shell date +"%Y/%m/%d %H:%M")
BUILD_COMMIT=$(shell git rev-parse HEAD)

export PATH := $(PATH):$(shell go env GOPATH)/bin

BIN_PATH=./bin/shortener
SRC_PATH=./internal
VET_TOOL=./bin/statictest

TEST_BIN=./bin/shortenertestbeta
T_BIN_FLAG=-binary-path=$(BIN_PATH)
T_SRC_FLAG=-source-path=$(SRC_PATH)
T_DSN_FLAG=-database-dsn="postgres://test:test@localhost/url_shortener_test"

gofmt:
	gofmt -s -l . 	

tests:
	go test -race ./...

vet:
	go vet -vettool=$$(which $(VET_TOOL)) ./... 

staticcheck:
	$(GOBIN)/staticcheck ./...

codecov:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

golangci-lint:
	$(GOBIN)/golangci-lint run ./...	

buildlint:
	go build -o=bin/staticlint cmd/staticlint/main.go 

lint: buildlint
	bin/staticlint ./...
	
runver:
	go run -ldflags \
		"-X main.buildVersion=$(BUILD_VERSION) -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(BUILD_COMMIT)' "\
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

test3:
	$(TEST_BIN) -test.v -test.run=^TestIteration3$$ $(T_SRC_FLAG)

test4:
	$(TEST_BIN) -test.v -test.run=^TestIteration4$$ $(T_BIN_FLAG) -server-port="9090"

test5:
	$(TEST_BIN) -test.v -test.run=^TestIteration5$$ $(T_BIN_FLAG) -server-port="9090"

test6:
	$(TEST_BIN) -test.v -test.run=^TestIteration6$$ $(T_SRC_FLAG)

test7:
	$(TEST_BIN) -test.v -test.run=^TestIteration7$$ $(T_BIN_FLAG) $(T_SRC_FLAG)

test8:
	$(TEST_BIN) -test.v -test.run=^TestIteration8$$ $(T_BIN_FLAG) 

test9:
	$(TEST_BIN) -test.v -test.run=^TestIteration9$$ $(T_BIN_FLAG) $(T_SRC_FLAG) -file-storage-path="./bin/file-store-test.json"

test10:
	$(TEST_BIN) -test.v -test.run=^TestIteration10$$ $(T_BIN_FLAG) $(T_SRC_FLAG) $(T_DSN_FLAG)

test11:
	$(TEST_BIN) -test.v -test.run=^TestIteration11$$ $(T_BIN_FLAG) $(T_DSN_FLAG)

test12:
	$(TEST_BIN) -test.v -test.run=^TestIteration12$$ $(T_BIN_FLAG) $(T_DSN_FLAG)

test13:
	$(TEST_BIN) -test.v -test.run=^TestIteration13$$ $(T_BIN_FLAG) $(T_DSN_FLAG)

test14:
	$(TEST_BIN) -test.v -test.run=^TestIteration14$$ $(T_BIN_FLAG) $(T_DSN_FLAG)

test15:
	$(TEST_BIN) -test.v -test.run=^TestIteration15$$ $(T_BIN_FLAG) $(T_DSN_FLAG)

alltests: tests build test1 test2 test3 test4 test5 test6 test7 test8 test9 test10 test11 test12 test13 test14 test15
	@echo "all tests - PASS"

# gRPC
protoc: 
	protoc --go_out=gen/go --go_opt=paths=source_relative \
	--go-grpc_out=gen/go --go-grpc_opt=paths=source_relative \
	proto/v1/shortener.proto

# run client rpc test
pbclient:
	go run cmd/grpcclient/main.go

# Error 403
# User Not Authorized
# This email isn't authorized to access the Buf Schema Registry
# не смог я получить доступ к сайту ни с каким vpn
# скопировал фалы локально
bufgen:
	buf generate --path proto