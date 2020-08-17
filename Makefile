PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go

$(PROTOC_GEN_GO):
	go get -u github.com/golang/protobuf/protoc-gen-go


.PHONY: proto
proto: $(PROTOC_GEN_GO)
	 protoc -I apigrpc/protos/ apigrpc/protos/kiddy.proto --go_out=plugins=grpc:apigrpc/


.PHONY: build
build:
	CGO_ENABLED=0 go build -o kiddy .


.PHONY: tests
tests:
	go test ./...


.PHONY: run
run:
	docker-compose up --build -d


.PHONY: stop
stop:
	docker-compose down


.PHONY: lint
lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.30.0

	golangci-lint run
