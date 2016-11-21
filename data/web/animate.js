'use strict';

(function() {
  // Just in case polyfill from:
  // https://www.paulirish.com/2011/requestanimationframe-for-smart-animating/
  var lastTime = 0;
  var vendors = ['webkit', 'moz'];

  for(var x = 0; x < vendors.length && !window.requestAnimationFrame; ++x) {
    window.requestAnimationFrame = window[vendors[x]+'RequestAnimationFrame'];
    window.cancelAnimationFrame = window[vendors[x]+'CancelAnimationFrame']
      || window[vendors[x]+'CancelRequestAnimationFrame'];
  }

  if (!window.requestAnimationFrame) {
    window.requestAnimationFrame = function(callback, element) {
      var currTime = new Date().getTime();
      var timeToCall = Math.max(0, 16 - (currTime - lastTime));
      var id = window.setTimeout(function() {
        callback(currTime + timeToCall);
      }, timeToCall);
      lastTime = currTime + timeToCall;
      return id;
    };
  }

  if (!window.cancelAnimationFrame) {
    window.cancelAnimationFrame = function(id) {
      clearTimeout(id);
    };
  }
}());

nmux.animate = new (function() {
  var self = this;
  var raf = window.requestAnimationFrame;
  var callbacks = {};
  var lastT = 0;

  function tick(t) {
    var name, item;

    raf(tick);

    for (name in callbacks) {
      item = callbacks[name];
      if (item.delay && t < item.delay) {
        continue;
      }

      if (t - item.last >= item.interval) {
        item.last = t;
        item.counter++;
        item.callback(t, item.counter);

        if (item.count > 0 && item.counter >= item.count) {
          self.remove(name);
        }
      }
    }

    lastT = t;
  }

  raf(tick);

  self.remove = function(name) {
    var item = callbacks[name];
    if (item) {
      delete callbacks[name];
    }
    return item;
  };

  self.add = function(name, cb, interval, delay, count) {
    if (!isNaN(delay)) {
      delay = parseInt(delay, 10);
      delay = lastT + delay;
    } else {
      delay = 0;
    }

    if (!isNaN(count)) {
      count = parseInt(count, 10);
    } else {
      count = 0;
    }

    callbacks[name] = {
      'callback': cb,
      'delay': delay,
      'count': count,
      'counter': 0,
      'interval': interval,
      'last': 0,
    };
  };
})();
