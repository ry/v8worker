#ifdef __cplusplus
extern "C" {
#endif


struct worker_s;
typedef struct worker_s worker;

const char* worker_version();

void v8_init();

worker* worker_new(int table_index);

// returns nonzero on error
// get error from worker_last_exception
int worker_load(worker* w, char* name_s, char* source_s);

const char* worker_last_exception(worker* w);

int worker_send(worker* w, const char* msg);
const char* worker_send_sync(worker* w, const char* msg);

void worker_dispose(worker* w);
void worker_terminate_execution(worker* w);

#ifdef __cplusplus
} // extern "C"
#endif
