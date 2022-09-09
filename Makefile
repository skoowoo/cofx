.PHONY: all build clean install
all:
	cofunc run make.flowl -e BUILD=true -e TEST=true

build:
	cofunc run make.flowl -e BUILD=true 

test:
	cofunc run make.flowl -e TEST=true 

clean:
	rm -rf bin/

first:
	go generate ./...
	go build -o bin/cofunc github.com/cofunclabs/cofunc/cmd/cofunc