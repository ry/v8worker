/*global $send*/

'use strict';

$send('module1 __dirname = ' + __dirname);
$send('module1 __filename = ' + __filename);

var module2 = require('module-2');

for (var i = 0; i < 10; i++) {
  $send(module2.saySomething(i));
}

if (require.main === module) {
  $send("module1 is main");
}

module2.longRunning(function(msg) {
  $send(msg);
});
