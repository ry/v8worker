#include <stdio.h>
#include <assert.h>
#include <stdlib.h>
#include <string.h>
#include <string>
#include "v8.h"
#include "libplatform/libplatform.h"
#include "binding.h"

using namespace v8;

struct worker_s {
  int x;
  void* data;
  worker_recv_cb cb;
  worker_recv_sync_cb sync_cb;
  Isolate* isolate;
  std::string last_exception;
  Persistent<Function> recv;
  Persistent<Context> context;
  Persistent<Function> recv_sync_handler;
};

// Extracts a C string from a V8 Utf8Value.
const char* ToCString(const String::Utf8Value& value) {
  return *value ? *value : "<string conversion failed>";
}

class ArrayBufferAllocator : public ArrayBuffer::Allocator {
 public:
  virtual void* Allocate(size_t length) {
    void* data = AllocateUninitialized(length);
    return data == NULL ? data : memset(data, 0, length);
  }
  virtual void* AllocateUninitialized(size_t length) { return malloc(length); }
  virtual void Free(void* data, size_t) { free(data); }
};


// Exception details will be appended to the first argument.
std::string ExceptionString(Isolate* isolate, TryCatch* try_catch) {
  std::string out;
  size_t scratchSize = 20;
  char scratch[scratchSize]; // just some scratch space for sprintf

  HandleScope handle_scope(isolate);
  String::Utf8Value exception(try_catch->Exception());
  const char* exception_string = ToCString(exception);

  Handle<Message> message = try_catch->Message();

  if (message.IsEmpty()) {
    // V8 didn't provide any extra information about this error; just
    // print the exception.
    out.append(exception_string);
    out.append("\n");
  } else {
    // Print (filename):(line number)
    String::Utf8Value filename(message->GetScriptOrigin().ResourceName());
    const char* filename_string = ToCString(filename);
    int linenum = message->GetLineNumber();

    snprintf(scratch, scratchSize, "%i", linenum);
    out.append(filename_string);
    out.append(":");
    out.append(scratch);
    out.append("\n");

    // Print line of source code.
    String::Utf8Value sourceline(message->GetSourceLine());
    const char* sourceline_string = ToCString(sourceline);

    out.append(sourceline_string);
    out.append("\n");

    // Print wavy underline (GetUnderline is deprecated).
    int start = message->GetStartColumn();
    for (int i = 0; i < start; i++) {
      out.append(" ");
    }
    int end = message->GetEndColumn();
    for (int i = start; i < end; i++) {
      out.append("^");
    }
    out.append("\n");
    String::Utf8Value stack_trace(try_catch->StackTrace());
    if (stack_trace.length() > 0) {
      const char* stack_trace_string = ToCString(stack_trace);
      out.append(stack_trace_string);
      out.append("\n");
    } else {
      out.append(exception_string);
      out.append("\n");
    }
  }
  return out;
}


extern "C" {
#include "_cgo_export.h"

void go_recv_cb(const char* msg, void* data) {
  recvCb((char*)msg, data);
}

const char* go_recv_sync_cb(const char* msg, void* data) {
  return recvSyncCb((char*)msg, data);
}

const char* worker_version() {
  return V8::GetVersion();
}

const char* worker_last_exception(worker* w) {
  return w->last_exception.c_str();
}

int worker_load(worker* w, char* name_s, char* source_s) {
  Locker locker(w->isolate);
  Isolate::Scope isolate_scope(w->isolate);
  HandleScope handle_scope(w->isolate);

  Local<Context> context = Local<Context>::New(w->isolate, w->context);
  Context::Scope context_scope(context);

  TryCatch try_catch;

  Local<String> name = String::NewFromUtf8(w->isolate, name_s);
  Local<String> source = String::NewFromUtf8(w->isolate, source_s);

  ScriptOrigin origin(name);

  Local<Script> script = Script::Compile(source, &origin);

  if (script.IsEmpty()) {
    assert(try_catch.HasCaught());
    w->last_exception = ExceptionString(w->isolate, &try_catch);
    return 1;
  }

  Handle<Value> result = script->Run();

  if (result.IsEmpty()) {
    assert(try_catch.HasCaught());
    w->last_exception = ExceptionString(w->isolate, &try_catch);
    return 2;
  }

  return 0;
}

void Print(const FunctionCallbackInfo<Value>& args) {
  bool first = true;
  for (int i = 0; i < args.Length(); i++) {
    HandleScope handle_scope(args.GetIsolate());
    if (first) {
      first = false;
    } else {
      printf(" ");
    }
    String::Utf8Value str(args[i]);
    const char* cstr = ToCString(str);
    printf("%s", cstr);
  }
  printf("\n");
  fflush(stdout);
}

// sets the recv callback.
void Recv(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  worker* w = (worker*)isolate->GetData(0);
  assert(w->isolate == isolate);

  HandleScope handle_scope(isolate);

  Local<Context> context = Local<Context>::New(w->isolate, w->context);
  Context::Scope context_scope(context);

  Local<Value> v = args[0];
  assert(v->IsFunction());
  Local<Function> func = Local<Function>::Cast(v);

  w->recv.Reset(isolate, func);
}

void RecvSync(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  worker* w = (worker*)isolate->GetData(0);
  assert(w->isolate == isolate);

  HandleScope handle_scope(isolate);

  Local<Context> context = Local<Context>::New(w->isolate, w->context);
  Context::Scope context_scope(context);

  Local<Value> v = args[0];
  assert(v->IsFunction());
  Local<Function> func = Local<Function>::Cast(v);

  w->recv_sync_handler.Reset(isolate, func);
}

// Called from javascript. Must route message to golang.
void Send(const FunctionCallbackInfo<Value>& args) {
  std::string msg;
  worker* w = NULL;
  {
    Isolate* isolate = args.GetIsolate();
    w = static_cast<worker*>(isolate->GetData(0));
    assert(w->isolate == isolate);

    Locker locker(w->isolate);
    HandleScope handle_scope(isolate);

    Local<Context> context = Local<Context>::New(w->isolate, w->context);
    Context::Scope context_scope(context);

    Local<Value> v = args[0];
    assert(v->IsString());

    String::Utf8Value str(v);
    msg = ToCString(str);
  }

  // XXX should we use Unlocker?
  w->cb(msg.c_str(), w->data);
}

// Called from javascript using $request.
// Must route message (string) to golang and send back message (string) as return value.
void SendSync(const FunctionCallbackInfo<Value>& args) {
  std::string msg;
  worker* w = NULL;
  {
    Isolate* isolate = args.GetIsolate();
    w = static_cast<worker*>(isolate->GetData(0));
    assert(w->isolate == isolate);

    Locker locker(w->isolate);
    HandleScope handle_scope(isolate);

    Local<Context> context = Local<Context>::New(w->isolate, w->context);
    Context::Scope context_scope(context);

    Local<Value> v = args[0];
    assert(v->IsString());

    String::Utf8Value str(v);
    msg = ToCString(str);
  }
  const char* returnMsg = w->sync_cb(msg.c_str(), w->data);
  Local<String> returnV = String::NewFromUtf8(w->isolate, returnMsg);
  args.GetReturnValue().Set(returnV);
}

// Called from golang. Must route message to javascript lang.
// non-zero return value indicates error. check worker_last_exception().
int worker_send(worker* w, const char* msg) {
  Locker locker(w->isolate);
  Isolate::Scope isolate_scope(w->isolate);
  HandleScope handle_scope(w->isolate);

  Local<Context> context = Local<Context>::New(w->isolate, w->context);
  Context::Scope context_scope(context);

  TryCatch try_catch;

  Local<Function> recv = Local<Function>::New(w->isolate, w->recv);
  if (recv.IsEmpty()) {
    w->last_exception = "$recv not called";
    return 1;
  }

  Local<Value> args[1];
  args[0] = String::NewFromUtf8(w->isolate, msg);

  assert(!try_catch.HasCaught());

  recv->Call(context->Global(), 1, args);

  if (try_catch.HasCaught()) {
    w->last_exception = ExceptionString(w->isolate, &try_catch);
    return 2;
  }

  return 0;
}

// Called from golang. Must route message to javascript lang.
// It will call the $recv_sync_handler callback function and return its string value.
const char* worker_send_sync(worker* w, const char* msg) {
  std::string out;
  Locker locker(w->isolate);
  Isolate::Scope isolate_scope(w->isolate);
  HandleScope handle_scope(w->isolate);

  Local<Context> context = Local<Context>::New(w->isolate, w->context);
  Context::Scope context_scope(context);

  Local<Function> recv_sync_handler = Local<Function>::New(w->isolate, w->recv_sync_handler);
  if (recv_sync_handler.IsEmpty()) {
    out.append("err: $recvSync not called");
    return out.c_str();
  }

  Local<Value> args[1];
  args[0] = String::NewFromUtf8(w->isolate, msg);
  Local<Value> response_value = recv_sync_handler->Call(context->Global(), 1, args);

  if (response_value->IsString()) {
    String::Utf8Value response(response_value->ToString());
    out.append(*response);
  } else {
    out.append("err: non-string return value");
  }
  return out.c_str();
}

static ArrayBufferAllocator array_buffer_allocator;

void v8_init() {
  V8::Initialize();

  Platform* platform = platform::CreateDefaultPlatform();
  V8::InitializePlatform(platform);

  V8::SetArrayBufferAllocator(&array_buffer_allocator);
}

worker* worker_new(worker_recv_cb cb, worker_recv_sync_cb sync_cb, void* data) {
  Isolate* isolate = Isolate::New();
  Locker locker(isolate);
  Isolate::Scope isolate_scope(isolate);
  HandleScope handle_scope(isolate);

  worker* w = new(worker);
  w->isolate = isolate;
  w->isolate->SetCaptureStackTraceForUncaughtExceptions(true);
  w->isolate->SetData(0, w);
  w->data = data;
  w->cb = cb;
  w->sync_cb = sync_cb;

  Local<ObjectTemplate> global = ObjectTemplate::New(w->isolate);

  global->Set(String::NewFromUtf8(w->isolate, "$print"),
              FunctionTemplate::New(w->isolate, Print));

  global->Set(String::NewFromUtf8(w->isolate, "$recv"),
              FunctionTemplate::New(w->isolate, Recv));

  global->Set(String::NewFromUtf8(w->isolate, "$send"),
              FunctionTemplate::New(w->isolate, Send));

  global->Set(String::NewFromUtf8(w->isolate, "$sendSync"),
              FunctionTemplate::New(w->isolate, SendSync));

  global->Set(String::NewFromUtf8(w->isolate, "$recvSync"),
              FunctionTemplate::New(w->isolate, RecvSync));

  Local<Context> context = Context::New(w->isolate, NULL, global);
  w->context.Reset(w->isolate, context);
  //context->Enter();

  return w;
}

void worker_dispose(worker* w) {
  w->isolate->Dispose();
  delete(w);
}

}
