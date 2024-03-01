#include <node_api.h>
#include <sys/ipc.h>
#include <sys/shm.h>
#include <iostream>

napi_value CreateSharedMemory(napi_env env, napi_callback_info info)
{
  size_t shm_size = 1024;
  key_t key = 1234;

  int shmid = shmget(key, shm_size, IPC_CREAT | 0666);
  void *shm = shmat(shmid, nullptr, 0);

  if (shm == (void *)-1)
  {
    napi_throw_error(env, nullptr, "Unable to attach to shared memory.");
    return nullptr;
  }

  napi_value result;
  napi_create_int32(env, shmid, &result);
  return result;
}

napi_value Init(napi_env env, napi_value exports)
{
  napi_value fn;
  napi_create_function(env, nullptr, 0, CreateSharedMemory, nullptr, &fn);
  napi_set_named_property(env, exports, "createSharedMemory", fn);
  return exports;
}

NAPI_MODULE(NODE_GYP_MODULE_NAME, Init)
