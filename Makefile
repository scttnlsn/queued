PREFIX=/usr/local
BINDIR=${PREFIX}/bin

all: build/queued

build:
	mkdir build

build/queued: build $(wildcard queued.go queued/*.go)
	go get -d -tags=${BUILD_TAGS}
	go build -o build/queued -tags=${BUILD_TAGS}

clean:
	rm -rf build

install: build/queued
	install -m 755 -d ${BINDIR}
	install -m 755 build/queued ${BINDIR}/queued

uninstall:
	rm ${BINDIR}/queued

test:
	go get -d -tags=${BUILD_TAGS}
	cd queued; go test -tags=${BUILD_TAGS}

.PHONY: install uninstall clean all test
