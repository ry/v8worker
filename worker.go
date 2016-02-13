package v8worker

/*
#cgo CXXFLAGS: -std=c++11
#cgo pkg-config: v8.pc
#include <stdlib.h>
#include "binding.h"
*/
import "C"
import (
	"errors"
	"runtime"
	"strconv"
	"sync"
	"unsafe"
)

var (
	// Don't init V8 more than once.
	initV8Once sync.Once

	scriptSequence       int = 0
	scriptSequenceLocker sync.Mutex
)

type workerTableIndex int

// To receive messages from javascript...
type ReceiveMessageCallback func(msg string)

// To send a message from javascript and synchronously return a string.
type ReceiveSyncMessageCallback func(msg string) string

// This is a golang wrapper around a single V8 Isolate.
type Worker struct {
	cWorker *C.worker
}

// ScriptOrigin represents V8 class – see http://v8.paulfryzel.com/docs/master/classv8_1_1_script_origin.html
type ScriptOrigin struct {
	ScriptName            string
	LineOffset            int32
	ColumnOffset          int32
	IsSharedCrossOrigin   bool
	ScriptId              int32
	IsEmbedderDebugScript bool
	SourceMapURL          string
	IsOpaque              bool
}

// Version return the V8 version E.G. "4.3.59"
func Version() string {
	return C.GoString(C.worker_version())
}

//export recvCb
func recvCb(msg_s *C.char, ptr unsafe.Pointer) {
	msg := C.GoString(msg_s)
	f := *(*ReceiveMessageCallback)(ptr)
	f(msg)
}

//export recvSyncCb
func recvSyncCb(msg_s *C.char, ptr unsafe.Pointer) *C.char {
	msg := C.GoString(msg_s)
	f := *(*ReceiveSyncMessageCallback)(ptr)
	return C.CString(f(msg))
}

// New creates a new worker, which corresponds to a V8 isolate. A single threaded
// standalone execution context.
func New(cb ReceiveMessageCallback, sync_cb ReceiveSyncMessageCallback) *Worker {
	worker := &Worker{}

	initV8Once.Do(func() {
		C.v8_init()
	})

	worker.cWorker = C.worker_new(unsafe.Pointer(&cb), unsafe.Pointer(&sync_cb))
	runtime.SetFinalizer(worker, func(final_worker *Worker) {
		C.worker_dispose(final_worker.cWorker)
	})
	return worker
}

// Load loads and executes a javascript file with the filename specified by
// scriptName and the contents of the file specified by the param code.
func (w *Worker) Load(scriptName string, code string) error {
	return w.LoadWithOptions(&ScriptOrigin{ScriptName: scriptName}, code)
}

// LoadWithOptions loads and executes a javascript file with the ScriptOrigin specified by
// origin and the contents of the file specified by the param code.
func (w *Worker) LoadWithOptions(origin *ScriptOrigin, code string) error {
	cCode := C.CString(code)

	if origin == nil {
		origin = new(ScriptOrigin)
	}
	if origin.ScriptName == "" {
		origin.ScriptName = nextScriptName()
	}
	cScriptName := C.CString(origin.ScriptName)
	cLineOffset := C.int(origin.LineOffset)
	cColumnOffset := C.int(origin.ColumnOffset)
	cIsSharedCrossOrigin := C.bool(origin.IsSharedCrossOrigin)
	cScriptId := C.int(origin.ScriptId)
	cIsEmbedderDebugScript := C.bool(origin.IsEmbedderDebugScript)
	cSourceMapURL := C.CString(origin.SourceMapURL)
	cIsOpaque := C.bool(origin.IsOpaque)

	defer C.free(unsafe.Pointer(cScriptName))
	defer C.free(unsafe.Pointer(cCode))
	defer C.free(unsafe.Pointer(cSourceMapURL))

	r := C.worker_load(w.cWorker, cCode, cScriptName, cLineOffset, cColumnOffset, cIsSharedCrossOrigin, cScriptId, cIsEmbedderDebugScript, cSourceMapURL, cIsOpaque)
	if r != 0 {
		errStr := C.GoString(C.worker_last_exception(w.cWorker))
		return errors.New(errStr)
	}
	return nil
}

// Send sends a message to a worker. The $recv callback in js will be called.
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

// TerminateExecution terminates execution of javascript
func (w *Worker) TerminateExecution() {
	C.worker_terminate_execution(w.cWorker)
}

func nextScriptName() string {
	scriptSequenceLocker.Lock()
	seq := scriptSequence
	scriptSequence++
	scriptSequenceLocker.Unlock()
	return "VM" + strconv.Itoa(seq)
}
