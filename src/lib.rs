use std::env;

use napi::bindgen_prelude::*;
use napi_derive::napi;

#[napi]
fn fibonacci(n: u32) -> u32 {
    match n {
        1 | 2 => 1,
        _ => fibonacci(n - 1) + fibonacci(n - 2),
    }
}

#[napi]
fn get_cwd<T: Fn(String) -> Result<()>>(callback: T) {
    callback(env::current_dir().unwrap().to_string_lossy().to_string()).unwrap();
}

#[napi]
fn test_callback<T>(callback: T)
where
    T: Fn(String) -> Result<()>,
{
}
