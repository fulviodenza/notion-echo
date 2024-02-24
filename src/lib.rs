use napi_derive::napi;
use teloxide::prelude::*;

#[napi]
async fn dice() {
    println!("Starting notion-echo bot...");

    let bot = Bot::from_env();

    teloxide::repl(bot, |bot: Bot, msg: Message| async move {
        bot.send_message(msg.chat.id, "Note saved to Notion")
            .await?;
        Ok(())
    })
    .await;
}

// fn access_shared_buffer(mut cx: FunctionContext) -> Result<JsUndefined> {
//     let buffer = cx.argument::<JsArrayBuffer>(0)?;
//     let slice = cx.borrow(&buffer, |data| {
//         let slice = data.as_slice::<u8>();
//     });
//     Ok(cx.undefined())
// }
