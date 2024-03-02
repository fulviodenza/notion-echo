use napi::{Error, Result};
use napi_derive::napi;
use shared_memory::*;
use std::error::Error as StdError;
use std::process::Command;
use std::result::Result as StdResult;
use teloxide::prelude::*;

#[napi]
async fn start() {
    println!("Starting notion-echo bot...");

    let id = create_shared_memory().unwrap();
    let bot = Bot::from_env();

    teloxide::repl(bot, move |bot: Bot, msg: Message| {
        let shmid = id.clone();
        println!("shmid {:?}", shmid);

        async move {
            if let Some(text) = msg.text() {
                let _ = write_to_shared_memory(&shmid, text).await;

                let send_result = bot.send_message(msg.chat.id, "Note saved to Notion").await;

                if let Err(e) = send_result {
                    println!("Error sending message: {}", e);
                }
            }
            Ok(())
        }
    })
    .await;
}

fn create_shared_memory() -> Result<String> {
    let shm = ShmemConf::new().size(1024).create().map_err(|e| {
        Error::new(
            napi::Status::GenericFailure,
            format!("Unable to create shared memory segment: {}", e),
        )
    })?;

    let _ = set_permissions(shm.get_os_id().to_owned().as_str(), "666").unwrap();
    Ok(shm.get_os_id().to_string())
}

async fn write_to_shared_memory(shmid: &str, data: &str) -> StdResult<(), Box<dyn StdError>> {
    println!("shared memory id {:?}", shmid);
    match ShmemConf::new().os_id(shmid).open() {
        Ok(mut shm) => {
            let shm_slice = unsafe { shm.as_slice_mut() };

            // Ensure that the data to write does not
            // exceed the size of the shared memory.
            if data.len() <= shm_slice.len() {
                shm_slice[..data.len()].copy_from_slice(data.as_bytes());
                println!("wrote {:?}", data);
                Ok(())
            } else {
                Err("Data exceeds shared memory size".into())
            }
        }
        Err(e) => {
            println!("Failed to open shared memory: {:?}", e);
            Err(Box::new(e))
        }
    }
}

fn set_permissions(shm_path: &str, permissions: &str) -> StdResult<(), Box<dyn StdError>> {
    Command::new("chmod")
        .arg(permissions)
        .arg(shm_path)
        .status()?;
    Ok(())
}
