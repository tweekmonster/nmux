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

  function store(name, value) {
    if (value === null) {
      return localStorage.removeItem(name);
    }

    if (arguments.length > 1) {
      return localStorage.setItem(name, JSON.stringify(value));
    }

    try {
      return JSON.parse(localStorage.getItem(name));
    } catch (e) {}

    return null;
  }

  var firstRun = true;
  var debug = store('nmux.debug') || false;

  function messageHandler(e) {
    var buf = new Reader(e.data);
    var op;

    scr.hideCursor();

    if (debug) {
      console.groupCollapsed('Payload (' + buf.remaining() + ')');
    }

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
          scr.setSize(buf.eint32(), buf.eint32());
          break;

        case nmux.OpPalette:
          var id, a, fg, bg, sp, len = buf.eint32();;
          while (len > 0) {
            id = buf.eint32();
            a = buf.uint8();
            fg = buf.eint32();
            bg = buf.eint32();
            sp = buf.eint32();
            scr.setPalette(id, a, fg, bg, sp);
            len--;
          }
          break;

        case nmux.OpStyle:
          scr.setAttributes(buf.eint32());
          break;

        case nmux.OpPut:
          var index = buf.eint32();
          var len = buf.eint32();
          var str = '';

          while (len > 0) {
            str += String.fromCharCode(buf.eint32());
            len--;
          }

          scr.renderText(str, index);
          break;

        case nmux.OpPutRep:
          var index = buf.eint32();
          var len = buf.eint32();
          var c = String.fromCharCode(buf.eint32());

          scr.renderRepeatedText(c, index, len);
          break;

        case nmux.OpScroll:
          var tmpBg = buf.uint24();
          var delta = buf.int16();
          var top = buf.eint32();
          var bottom = buf.eint32();
          var left = buf.eint32();
          var right = buf.eint32();

          scr.scroll(tmpBg, delta, left, top, right, bottom);
          break;

        case nmux.OpClear:
          var id, a, fg, bg, sp;
          id = buf.eint32();
          a = buf.uint8();
          fg = buf.uint24();
          bg = buf.uint24();
          sp = buf.uint24();
          scr.clear(id, a, fg, bg, sp);
          break;

        case nmux.OpFlush:
          scr.flush();

          var mode = buf.eint32();
          var cx = buf.eint32();
          var cy = buf.eint32();
          var id = buf.eint32();
          var c = buf.eint32();

          scr.setCursor(mode, cx, cy, id, c);
          scr.showCursor();
          break;

        default:
          console.log('Unknown Op', op);
      }


      if (debug) {
        buf.dumpLastRead(nmux['o' + op]);
      }
    }

    if (debug) {
      console.groupEnd();
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

    // Meta-Shift-D enables debug.
    if (key === '<D-D>') {
      debug = !debug;
      store('nmux.debug', debug);
      scr.setDebug(debug);
      return;
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

    if (!sock.send(data)) {
      sock.connect();
    }
  }

  var sockInit = false;
  sock = new nmux.socket(socketURL('/nmux'));
  sock.binaryType = 'arraybuffer';
  sock.addEventListener('open', function() {
    resize();
    scr.setDebug(debug);

    if (!sockInit) {
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
      sockInit = true;
    }
  });

  window.sock = sock;
});
