package v8worker

/*
#cgo CXXFLAGS: -std=c++11
#cgo pkg-config: v8.pc
#include <stdlib.h>
#include "binding.h"
*/
import "C"
import "errors"

import "unsafe"
import "sync"

// To receive messages from javascript...
type ReceiveMessageCallback func(msg string)

// To send a message from javascript and synchronously return a string.
type ReceiveSyncMessageCallback func(msg string) string

// DiscardSendSync can be used in the worker constructor when you don't use the builtin $sendSync.
func DiscardSendSync(msg string) string { return "" }

// Don't init V8 more than once.
var initV8Once sync.Once

// This is a golang wrapper around a single V8 Isolate.
type Worker struct {
	cWorker     *C.worker
	cb          ReceiveMessageCallback
	recvSync_cb ReceiveSyncMessageCallback
}

// Return the V8 version E.G. "4.3.59"
func Version() string {
	return C.GoString(C.worker_version())
}

//export recvCb
func recvCb(msg_s *C.char, ptr unsafe.Pointer) {
	msg := C.GoString(msg_s)
	worker := (*Worker)(ptr)
	worker.cb(msg)
}

//export recvSyncCb
func recvSyncCb(msg_s *C.char, ptr unsafe.Pointer) *C.char {
	msg := C.GoString(msg_s)
	worker := (*Worker)(ptr)
	return_s := C.CString(worker.recvSync_cb(msg))
	return return_s
}

// Creates a new worker, which corresponds to a V8 isolate. A single threaded
// standalone execution context.
func New(cb ReceiveMessageCallback, recvSync_cb ReceiveSyncMessageCallback) *Worker {
	worker := &Worker{
		cb:          cb,
		recvSync_cb: recvSync_cb,
	}

	initV8Once.Do(func() {
		C.v8_init()
	})

	callback := C.worker_recv_cb(C.go_recv_cb)
	receiveSync_callback := C.worker_recvSync_cb(C.go_recvSync_cb)

	worker.cWorker = C.worker_new(callback, receiveSync_callback, unsafe.Pointer(worker))
	return worker
}

// Load and executes a javascript file with the filename specified by
// scriptName and the contents of the file specified by the param code.
func (w *Worker) Load(scriptName string, code string) error {
	scriptName_s := C.CString(scriptName)
	code_s := C.CString(code)
	defer C.free(unsafe.Pointer(scriptName_s))
	defer C.free(unsafe.Pointer(code_s))

	r := C.worker_load(w.cWorker, scriptName_s, code_s)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.cWorker))
		return errors.New(errStr)
	}
	return nil
}

// Sends a message to a worker. The $recv callback in js will be called.
func (w *Worker) Send(msg string) error {
	msg_s := C.CString(string(msg))
	defer C.free(unsafe.Pointer(msg_s))

	r := C.worker_send(w.cWorker, msg_s)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.cWorker))
		return errors.New(errStr)
	}

	return nil
}

// SendSync sends a message to a worker. The $recvSync callback in js will be called.
// That callback will return a string which is passed to golang and used as the return value of SendSync.
func (w *Worker) SendSync(msg string) string {
	msg_s := C.CString(string(msg))
	defer C.free(unsafe.Pointer(msg_s))

	svalue := C.worker_sendSync(w.cWorker, msg_s)
	return C.GoString(svalue)
}
