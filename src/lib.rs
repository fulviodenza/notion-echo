use napi::{Error, Result};
use napi_derive::napi;
use shared_memory::*;
use teloxide::prelude::*;

#[napi]
async fn start() {
    println!("Starting notion-echo bot...");

    let bot = Bot::from_env();

    teloxide::repl(bot, |bot: Bot, msg: Message| async move {
        // TODO: handle messages
        bot.send_message(msg.chat.id, "Note saved to Notion")
            .await?;
        Ok(())
    })
    .await;
}

#[napi]
async fn create_shared_memory() -> Result<String> {
    let shm = ShmemConf::new().size(1024).create().map_err(|e| {
        Error::new(
            napi::Status::GenericFailure,
            format!("Unable to create shared memory segment: {}", e),
        )
    })?;

    Ok(shm.get_os_id().to_string())
}
