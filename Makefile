.PHONY: all build clean install
all:
	cofx run make.flowl -e BUILD=true -e TEST=true

build:
	cofx run make.flowl -e BUILD=true 

test:
	cofx run make.flowl -e TEST=true 
release:
	./release.sh

clean:
	rm -rf bin/ .tmp/