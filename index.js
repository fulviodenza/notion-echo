const bot = require("./index.node");

async function run() {
    const shmid = await bot.createSharedMemory();
    console.log(`Shared memory segment ID: ${shmid}`);
}

run()
bot.start()
