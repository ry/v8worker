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

type workerTableIndex int

var workerTableLock sync.Mutex

// This table will store all pointers to all active workers. Because we can't safely
// pass pointers to Go objects to C, we instead pass a key to this table.
var workerTable = make(map[workerTableIndex]*worker)

// Keeps track of the last used table index. Incremeneted when a worker is created.
var workerTableNextAvailable workerTableIndex = 0

// To receive messages from javascript...
type ReceiveMessageCallback func(msg string)

// To send a message from javascript and synchronously return a string.
type ReceiveSyncMessageCallback func(msg string) string

// Don't init V8 more than once.
var initV8Once sync.Once

// Internal worker struct which is stored in the workerTable.
// Weak-ref pattern https://groups.google.com/forum/#!topic/golang-nuts/1ItNOOj8yW8/discussion
type worker struct {
	cWorker    *C.worker
	cb         ReceiveMessageCallback
	sync_cb    ReceiveSyncMessageCallback
	tableIndex workerTableIndex
}

// This is a golang wrapper around a single V8 Isolate.
type Worker struct {
	*worker
}

// Return the V8 version E.G. "4.3.59"
func Version() string {
	return C.GoString(C.worker_version())
}

func workerTableLookup(index workerTableIndex) *worker {
	workerTableLock.Lock()
	defer workerTableLock.Unlock()
	return workerTable[index]
}

//export recvCb
func recvCb(msg_s *C.char, index workerTableIndex) {
	msg := C.GoString(msg_s)
	w := workerTableLookup(index)
	w.cb(msg)
}

//export recvSyncCb
func recvSyncCb(msg_s *C.char, index workerTableIndex) *C.char {
	msg := C.GoString(msg_s)
	w := workerTableLookup(index)
	return_s := C.CString(w.sync_cb(msg))
	return return_s
}

// Creates a new worker, which corresponds to a V8 isolate. A single threaded
// standalone execution context.
func New(cb ReceiveMessageCallback, sync_cb ReceiveSyncMessageCallback) *Worker {
	workerTableLock.Lock()
	w := &worker{
		cb:         cb,
		sync_cb:    sync_cb,
		tableIndex: workerTableNextAvailable,
	}

	workerTableNextAvailable++
	workerTable[w.tableIndex] = w
	workerTableLock.Unlock()

	initV8Once.Do(func() {
		C.v8_init()
	})

	w.cWorker = C.worker_new(C.int(w.tableIndex))

	externalWorker := &Worker{worker: w}
	
	return externalWorker
}

// Load and executes a javascript file with the filename specified by
// scriptName and the contents of the file specified by the param code.
func (w *Worker) Load(scriptName string, code string) error {
	scriptName_s := C.CString(scriptName)
	code_s := C.CString(code)
	defer C.free(unsafe.Pointer(scriptName_s))
	defer C.free(unsafe.Pointer(code_s))

	r := C.worker_load(w.worker.cWorker, scriptName_s, code_s)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.worker.cWorker))
		return errors.New(errStr)
	}
	return nil
}

// Sends a message to a worker. The $recv callback in js will be called.
func (w *Worker) Send(msg string) error {
	msg_s := C.CString(string(msg))
	defer C.free(unsafe.Pointer(msg_s))

	r := C.worker_send(w.worker.cWorker, msg_s)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.worker.cWorker))
		return errors.New(errStr)
	}

	return nil
}

// SendSync sends a message to a worker. The $recvSync callback in js will be called.
// That callback will return a string which is passed to golang and used as the return value of SendSync.
func (w *Worker) SendSync(msg string) string {
	msg_s := C.CString(string(msg))
	defer C.free(unsafe.Pointer(msg_s))

	svalue := C.worker_send_sync(w.worker.cWorker, msg_s)
	return C.GoString(svalue)
}

// Terminates execution of javascript
func (w *Worker) TerminateExecution() {
	C.worker_terminate_execution(w.worker.cWorker)
}

// Dispose worker and free memory
func (w *Worker) Dispose() {
	workerTableLock.Lock()
	delete(workerTable, w.tableIndex)
	workerTableLock.Unlock()
	C.worker_dispose(w.worker.cWorker)
}
