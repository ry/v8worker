#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

void go_recv_cb(const char* msg, int table_index);
const char* go_recv_sync_cb(const char* msg, int table_index);

struct worker_s;
typedef struct worker_s worker;
typedef void (*worker_recv_cb)(const char* msg, int table_index);
typedef const char* (*worker_recv_sync_cb)(const char* msg, int table_index);

const char* worker_version();

void v8_init();

worker* worker_new(worker_recv_cb cb, worker_recv_sync_cb sync_cb, int table_index);

// returns nonzero on error
// get error from worker_last_exception
int worker_load(worker* w, char* source_s, char* name_s, int line_offset_s, int column_offset_s, bool is_shared_cross_origin_s, int script_id_s, bool is_embedder_debug_script_s, char* source_map_url_s, bool is_opaque_s);

const char* worker_last_exception(worker* w);

int worker_send(worker* w, const char* msg);
const char* worker_send_sync(worker* w, const char* msg);

void worker_dispose(worker* w);
void worker_terminate_execution(worker* w);

#ifdef __cplusplus
} // extern "C"
#endif
