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
import "runtime"

type workerTableIndex int

var workerTableLock sync.Mutex

// This table will store all pointers to all active workers. Because we can't safely
// pass pointers to Go objects to C, we instead pass a key to this table.
var workerTable = make(map[workerTableIndex]*Worker)

// Keeps track of the last used table index. Incremeneted when a worker is created.
var workerTableNextAvailable workerTableIndex = 0

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
	cWorker    *C.worker
	cb         ReceiveMessageCallback
	sync_cb    ReceiveSyncMessageCallback
	tableIndex workerTableIndex
}

// Return the V8 version E.G. "4.3.59"
func Version() string {
	return C.GoString(C.worker_version())
}

func workerTableLookup(index workerTableIndex) *Worker {
	workerTableLock.Lock()
	defer workerTableLock.Unlock()
	return workerTable[index]
}

//export recvCb
func recvCb(msg_s *C.char, index workerTableIndex) {
	msg := C.GoString(msg_s)
	worker := workerTableLookup(index)
	worker.cb(msg)
}

//export recvSyncCb
func recvSyncCb(msg_s *C.char, index workerTableIndex) *C.char {
	msg := C.GoString(msg_s)
	worker := workerTableLookup(index)
	return_s := C.CString(worker.sync_cb(msg))
	return return_s
}

// Creates a new worker, which corresponds to a V8 isolate. A single threaded
// standalone execution context.
func New(cb ReceiveMessageCallback, sync_cb ReceiveSyncMessageCallback) *Worker {
	workerTableLock.Lock()
	worker := &Worker{
		cb:         cb,
		sync_cb:    sync_cb,
		tableIndex: workerTableNextAvailable,
	}

	workerTableNextAvailable++
	workerTable[worker.tableIndex] = worker
	workerTableLock.Unlock()

	initV8Once.Do(func() {
		C.v8_init()
	})

	callback := C.worker_recv_cb(C.go_recv_cb)
	receiveSync_callback := C.worker_recv_sync_cb(C.go_recv_sync_cb)

	worker.cWorker = C.worker_new(callback, receiveSync_callback, C.int(worker.tableIndex))
	runtime.SetFinalizer(worker, func(final_worker *Worker) {
		workerTableLock.Lock()
		delete(workerTable, final_worker.tableIndex)
		workerTableLock.Unlock()
		C.worker_dispose(final_worker.cWorker)
	})
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

	svalue := C.worker_send_sync(w.cWorker, msg_s)
	return C.GoString(svalue)
}
