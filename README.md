v8worker
========

Minimal golang binding to V8. This exposes a non-blocking message passing
interface to the V8 javascript engine. Go and JavaScript interact by sending
and receiving messages. V8 will block a thread (goroutine) only while it
computes javascript - it has no "syscalls" other than sending and receiving
messages to Go. There are only three built in functions exposed to javascript:
`$print(string)`, `$send(msg)`, and `$recv(callback)`. 

MIT License. Contributions welcome.

Build
-----

Run `make` to trigger a build of V8 and then do `make install`
which will trigger `go install`. V8 is statically linked. It's only been tested
on my OSX laptop so far - should be easily portable to linux tho.

`make test` to build/run tests. Or just `go test`.

Docs
----

From golang checkout the API here:
https://godoc.org/github.com/ry/v8worker

From Javascript you only have:
`$print(string)`
`$send(msg)`
`$recv(callback)`. 
See `worker_test.go` for example usage for now.



TODO
----
- more tests
- need ability to pass command line options to V8 when creating a worker (maybe before)
- way to kill worker
- get text of exception
