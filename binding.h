#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

struct worker_s;
typedef struct worker_s worker;

const char* worker_version();

void v8_init();

worker* worker_new(int worker_id);

// returns nonzero on error
// get error from worker_last_exception
int worker_load(worker* w, char* source_s, char* name_s, int line_offset_s, int column_offset_s, bool is_shared_cross_origin_s, int script_id_s, bool is_embedder_debug_script_s, char* source_map_url_s, bool is_opaque_s);

const char* worker_last_exception(worker* w);

int worker_send(worker* w, const char* msg);
const char* worker_send_sync(worker* w, const char* msg);

void worker_dispose(worker* w);
void worker_terminate_execution(worker* w);
void worker_low_memory_notification(worker* w);
bool worker_idle_notification_deadline(worker* w, double deadline_in_seconds);

#ifdef __cplusplus
} // extern "C"
#endif
