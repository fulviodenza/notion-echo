const bot = require("./index.node");

process.on('SIGINT', () => {
    console.log('Received SIGINT. Exiting...');
    process.exit(1);
});

async function run() {
    await bot.start();
}

run()
