DEPS = $(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

test:
	go test -v -timeout=60s .

integ: test
	INTEG_TESTS=yes go test -v -timeout=25s -run=Integ .

fuzz:
	go test -timeout=300s ./fuzzy
	
deps:
	go get -d -v ./...
	echo $(DEPS) | xargs -n1 go get -d

cov:
	INTEG_TESTS=yes gocov test github.com/readystock/raft | gocov-html > /tmp/coverage.html
	open /tmp/coverage.html


PROTOS_DIRECTORY = ./protos

protos:
	protoc -I=$(PROTOS_DIRECTORY) --go_out=plugins=grpc:./ $(PROTOS_DIRECTORY)/log.proto
	protoc -I=$(PROTOS_DIRECTORY) --go_out=plugins=grpc:./ $(PROTOS_DIRECTORY)/rpc_header.proto
	protoc -I=$(PROTOS_DIRECTORY) --go_out=plugins=grpc:./ $(PROTOS_DIRECTORY)/append_entries.proto
	protoc -I=$(PROTOS_DIRECTORY) --go_out=plugins=grpc:./ $(PROTOS_DIRECTORY)/request_vote.proto
	protoc -I=$(PROTOS_DIRECTORY) --go_out=plugins=grpc:./ $(PROTOS_DIRECTORY)/install_snapshot.proto
	protoc -I=$(PROTOS_DIRECTORY) --go_out=plugins=grpc:./ $(PROTOS_DIRECTORY)/service.proto

.PHONY: test cov integ deps protos
