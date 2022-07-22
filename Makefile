.PHONY: build clean install
build:
	flowl run build/make.flowl

clean:
	rm -rf bin/

install:
	rm ${HOME}/local/bin/flowl
	cp bin/flowl ${HOME}/local/bin/