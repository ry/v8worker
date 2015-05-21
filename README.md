v8worker
========

[![Join the chat at https://gitter.im/ry/v8worker](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/ry/v8worker?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Minimal golang binding to V8. This exposes a non-blocking message passing
interface to the V8 javascript engine. Go and JavaScript interact by sending
and receiving messages. V8 will block a thread (goroutine) only while it
computes javascript - it has no "syscalls" other than sending and receiving
messages to Go. There are only three built in functions exposed to javascript:
`$print(string)`, `$send(msg)`, and `$recv(callback)`. 

MIT License. Contributions welcome.

Build
-----

You will need chrome's `depot_tools` to build. Follow the instructions here
https://www.chromium.org/developers/how-tos/install-depot-tools

Run `make` to trigger a download and build of V8. `make install` will trigger
`go install`. V8 is statically linked. It's only been tested on my OSX laptop
and x64 linux. Should be portable with some difficulty to windows.

`make test` to build/run tests. Or just `go test`.

To build a debug version use `target=x64.debug make`

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
