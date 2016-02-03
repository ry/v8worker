v8worker
========

Minimal golang binding to V8. This exposes a non-blocking message passing
interface to the V8 javascript engine. Go and JavaScript interact by sending
and receiving messages. V8 will block a thread (goroutine) only while it
computes javascript - it has no "syscalls" other than sending and receiving
messages to Go. There are only a few built in functions exposed to javascript:
`$print(string)`, `$send(msg)`, `$recv(callback)`, `$sendSync(msg)`, and
`$recvSync(callback)` 

[A slightly out of date presentation on this project](https://docs.google.com/presentation/d/1RgGVgLuP93mPZ0lqHhm7TOpxZBI3TEdAJQZzFqeleAE/edit?usp=sharing)

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
`$sendSync(msg)`. 
`$recvSync(callback)`. 
See `worker_test.go` for example usage for now.



TODO
----
- need ability to pass command line options to V8 when creating a worker (maybe before)
- get text of exception
