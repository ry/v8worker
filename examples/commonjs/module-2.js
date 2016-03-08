/*global $send*/

'use strict';

$send('module2 __dirname = ' + __dirname);
$send('module2 __filename = ' + __filename);

var words = ['every', 'good', 'bird', 'does', 'fly',
  'alligators', 'can', 'eat', 'grape', 'burgers'
];

exports.saySomething = function(index) {
  return 'module2 says ' + words[index % 10] + '!';
};

exports.longRunning = function(callback) {
  for (var i = 0; i < Number.MAX_SAFE_INTEGER; i++) {
    if (i === Math.floor(Number.MAX_SAFE_INTEGER / 10000000)) {
      callback('long running job has finished (' + i + ')');
      return;
    }
  }
};
