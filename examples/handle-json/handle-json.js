/*global $recv,$send*/

'use strict';

$recv(function(msg) {
  var obj = JSON.parse(msg);
  obj.handled = true;
  $send(JSON.stringify(obj));
});
