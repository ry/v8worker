#ifdef __cplusplus
extern "C" {
#endif


void go_recv_cb(const char* msg, void* data);
void go_print_cb(const char* str, void* data);
void c_print_cb(const char* str, void* data);

struct worker_s;
typedef struct worker_s worker;
typedef void (*worker_recv_cb)(const char* msg, void* data);
typedef void (*worker_print_cb)(const char* str, void* data);

const char* worker_version();

void v8_init();

worker* worker_new(worker_recv_cb cb, worker_print_cb print, void* data);

// returns nonzero on error
// get error from worker_last_exception
int worker_load(worker* w, char* name_s, char* source_s);

const char* worker_last_exception(worker* w);

int worker_send(worker* w, const char* msg);

#ifdef __cplusplus
} // extern "C"
#endif
