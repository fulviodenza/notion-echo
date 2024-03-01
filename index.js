const sharedBuffer = require("./index.node");

async function run() {
    const shmid = await sharedBuffer.createSharedMemory();
    console.log(`Shared memory segment ID: ${shmid}`);
}

run()
