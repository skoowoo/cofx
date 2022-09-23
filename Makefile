.PHONY: all build clean install
all:
	cofx run make.flowl -e BUILD=true -e TEST=true

build:
	cofx run make.flowl -e BUILD=true 

test:
	cofx run make.flowl -e TEST=true 

clean:
	rm -rf bin/ .tmp/

first:
	go generate ./...
	go build -o bin/cofx github.com/cofxlabs/cofx/cmd/cofx