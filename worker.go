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

type Message string // just JSON for now...
type RecieveMessageCallback func(msg Message)

type Worker struct {
	cWorker *C.worker
	cb      RecieveMessageCallback
}

func Version() string {
	return C.GoString(C.worker_version())
}

//export recvCb
func recvCb(msg_s *C.char, ptr unsafe.Pointer) {
	msg := Message(C.GoString(msg_s))
	worker := (*Worker)(ptr)
	worker.cb(msg)
}

func New(cb RecieveMessageCallback) *Worker {
	worker := &Worker{
		cb: cb,
	}

	callback := C.worker_recv_cb(C.go_recv_cb)

	worker.cWorker = C.worker_new(callback, unsafe.Pointer(worker))
	return worker
}

func (w *Worker) Load(scriptName string, code string) error {
	scriptName_s := C.CString(scriptName)
	code_s := C.CString(code)
	r := C.worker_load(w.cWorker, scriptName_s, code_s)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.cWorker))
		return errors.New(errStr)
	}
	return nil
}

func (w *Worker) Send(msg Message) error {
	msg_s := C.CString(string(msg))
	defer C.free(unsafe.Pointer(msg_s))

	r := C.worker_send(w.cWorker, msg_s)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.cWorker))
		return errors.New(errStr)
	}

	return nil
}
