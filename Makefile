.PHONY: build clean install
build:
	cofunc run build/make.flowl

clean:
	rm -rf bin/

install:
	rm -f ${HOME}/local/bin/cofunc
	cp bin/cofunc ${HOME}/local/bin/