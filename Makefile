version?=4.3.59
target?=x64.debug # available: native, ia32.debug, ia32.release, x64.release
prefix?=src/v8worker

test: v8worker.test
	./v8worker.test

v8worker.test: v8.pc ${prefix}/*.go ${prefix}/*.cc ${prefix}/*.h
	GOPATH=`pwd` go test -c

install: v8.pc ${prefix}/*.go ${prefix}/*.cc ${prefix}/*.h
	GOPATH=`pwd` cd src/v8worker && go install

v8.pc:
	version=$(version) target=$(target) ./build.sh

clean:
	rm -f v8.pc v8worker.test

distclean: clean
	rm -rf v8-$(version)/out


.PHONY: install test clean distclean
