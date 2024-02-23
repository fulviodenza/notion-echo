const { dice } = require("./index.node");
var addon = require('bindings')('hello');

const sharedBuffer = new SharedArrayBuffer(1024);

console.log(addon.hello())
console.log(dice())
