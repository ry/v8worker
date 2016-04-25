version?=5.1-lkgr
target?=native # available: x64.debug, ia32.debug, ia32.release, x64.release

test: v8worker.test
	./v8worker.test

examples: examples/*/*.go examples/*/*.js
	cd examples/commonjs && go build && ./commonjs
	cd examples/handle-json && go build && ./handle-json

v8.pc: v8
	target=$(target) ./build.sh

v8:
	fetch --nohooks v8
	cd v8 && git checkout $(version) && gclient sync

v8worker.test: v8.pc *.go *.cc *.h
	go test -c

install: v8.pc *.go *.cc *.h
	go install


clean:
	rm -f v8.pc v8worker.test

distclean: clean
	rm -f .gclient .gclient_entries
	rm -rf v8/

.PHONY: install test clean distclean examples
