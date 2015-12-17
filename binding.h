#ifdef __cplusplus
extern "C" {
#endif


void go_recv_cb(const char* msg, void* data);
const char* go_recvSync_cb(const char* msg, void* data);

struct worker_s;
typedef struct worker_s worker;
struct context_s;
typedef struct context_s context;
typedef void (*worker_recv_cb)(const char* msg, void* data);
typedef const char* (*worker_recvSync_cb)(const char* msg, void* data);

const char* worker_version();
void v8_init();

worker* worker_new(void* data);
context* context_new(worker* w, worker_recv_cb cb, worker_recvSync_cb recvSync_cb);

// returns nonzero on error
// get error from worker_last_exception
int worker_load(worker* w, context *c, char* name_s, char* source_s);

const char* worker_last_exception(worker* w);

int worker_send(worker* w, context *c, const char* msg);
const char* worker_sendSync(worker* w, context *c, const char* msg);

void worker_dispose(worker* w);
void context_dispose(context* c);

#ifdef __cplusplus
} // extern "C"
#endif
