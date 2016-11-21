'use strict';

new (function() {
  var sock;
  var scr = new Screen();

  function socketURL(s) {
    var l = window.location;
    return (l.protocol === 'https:' ? 'wss://' : 'ws://') + l.hostname
      + ((l.port != 80 && l.port != 443) ? ':' + l.port : '') + s;
  }

  function debounce(func, wait, immediate) {
    var timeout;
    return function() {
      var self = this;
      var args = arguments;
      var later = function() {
        timeout = null;
        if (!immediate) {
          func.apply(self, args);
        }
      };

      var callNow = immediate && timeout;
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
      if (callNow) {
        func.apply(self, args);
      }
    };
  }

  function resize() {
    var c = scr.charSize();
    var w = Math.floor(document.body.offsetWidth / c[0]);
    var h = Math.floor(document.body.offsetHeight / c[1]);
    var payload = [
      nmux.OpResize,
      w >> 8, w & 0xff,
      h >> 8, h & 0xff,
    ];
    sock.send(new Uint8Array(payload));
  }

  var firstRun = true;

  function messageHandler(e) {
    var buf = new Reader(e.data);
    var op;

    scr.hideCursor();

    while (buf.remaining() > 0) {
      op = buf.uint8();

      if (firstRun && op != nmux.OpResize) {
        // Requires the first operation seen to be a resize to ensure we get a
        // full screen the first time.
        return;
      }


      switch(op) {
        case nmux.OpResize:
          firstRun = false;
          scr.setSize(buf.uint16(), buf.uint16());
          break;

        case nmux.OpStyle:
          var attrs = buf.uint8();
          var fg = buf.uint24();
          var bg = buf.uint24();
          var sp = buf.uint24();
          scr.setAttributes(attrs, fg, bg, sp);
          break;

        case nmux.OpPut:
          var index = buf.uint16();
          var len = buf.uint16();
          var str = '';

          while (len > 0) {
            str += String.fromCharCode(buf.uint16());
            len--;
          }

          scr.renderText(str, index);
          break;

        case nmux.OpPutRep:
          var index = buf.uint16();
          var len = buf.uint16();
          var c = String.fromCharCode(buf.uint16());

          scr.renderRepeatedText(c, index, len);
          break;

        case nmux.OpScroll:
          var tmpBg = buf.uint24();
          var delta = buf.int16();
          var top = buf.uint16();
          var bottom = buf.uint16();
          var left = buf.uint16();
          var right = buf.uint16();

          scr.scroll(tmpBg, delta, left, top, right, bottom);
          break;

        case nmux.OpClear:
          scr.clear();
          break;

        case nmux.OpFlush:
          scr.flush();

          var mode = buf.uint8();
          var cx = buf.uint16();
          var cy = buf.uint16();

          var c = buf.uint16();
          var a = buf.uint8();
          var fg = buf.uint24();
          var bg = buf.uint24();
          var sp = buf.uint24();

          scr.setCursor(mode, cx, cy, c, a, fg, bg, sp);
          scr.showCursor();
          break;

        default:
          console.log('Unknown Op', op);
      }
    }
  }

  function keyHandler(e) {
    e.preventDefault();
    e.stopPropagation();

    var key = '';

    if (e.constructor === MouseEvent) {
      if (e.type == 'mousemove' && e.buttons === 0) {
        return;
      }
      key = nmuxMouseKey.call(this, e);
    } else if (e.constructor === WheelEvent) {
      key = nmuxMouseWheel.call(this, e);
    } else {
      key = nmuxKey.call(this, e);
    }

    if (!key) {
      return;
    }

    var data = new ArrayBuffer(key.length + 1);
    var out = new Uint8Array(data);
    out[0] = nmux.OpKeyboard;
    for (var i = 0, l = key.length; i < l; i++) {
      out[i + 1] = key.charCodeAt(i);
    }

    sock.send(data);
  }

  // TODO: Retry on disconnect.
  sock = new WebSocket(socketURL('/nmux'));
  sock.binaryType = 'arraybuffer';
  sock.addEventListener('open', function() {
    resize();

    window.addEventListener('resize', debounce(resize, 200));
    window.addEventListener('keydown', keyHandler);
    scr.addEventListener('mousedown', keyHandler);
    scr.addEventListener('mouseup', keyHandler);
    scr.addEventListener('mousemove', keyHandler);
    scr.addEventListener('wheel', keyHandler);
    scr.addEventListener('contextmenu', function(e) {
      e.preventDefault();
      e.stopPropagation();
    });

    sock.addEventListener('message', messageHandler);
  });

  window.sock = sock;
});
