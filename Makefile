PREFIX=/usr/local
BINDIR=${PREFIX}/bin

all: build/kew

build:
	mkdir build

build/kew: build
	go build -o build/kew

clean:
	rm -rf build

install: build/kew
	install -m 755 -d ${BINDIR}
	install -m 755 build/kew ${BINDIR}/kew

uninstall:
	rm ${BINDIR}/kew

test:
	cd kew; go test

.PHONY: install uninstall clean all test