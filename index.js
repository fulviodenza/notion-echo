const { fibonacci } = require("./index.node");
var addon = require('bindings')('hello');

console.log(fibonacci(10))
console.log(addon.hello())