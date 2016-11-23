'use strict';

var Screen = function() {
  var self = this;

  // Tracked in another script.
  self.mouse = {
    'x': 0,
    'y': 0,
    'event': '',
  }

  var font = '';
  var charW = 0;
  var charH = 0;
  var charLH = 0;
  var charOffsetY = 0;
  var gridW = 0;
  var gridH = 0;
  var cursorX = 0;
  var cursorY = 0;
  var bgCache = {};
  var palette = {};
  var brush = null;
  var repeatCache = {};

  // Global X coordinate for drawing undercurls whose ends meet regardless of
  // what canvs they're drawn on.
  var gX = 0;

  var grid = new Uint8Array(12);

  new (function() {
    var g = document.createElement('span');
    g.setAttribute('class', 'glyph');
    g.style.display = 'inline-block';
    g.style.verticalAlign = 'baseline';
    document.body.appendChild(g);
    var f = window.getComputedStyle(g);

    font = f.fontSize + ' ' + f.fontFamily;

    g.textContent = 'X';
    charW = g.offsetWidth;
    charH = g.offsetHeight;

    g.style.fontStyle = 'italic';
    g.style.fontWeight = 'bold';
    charW = Math.max(charW, g.offsetWidth);
    charH = Math.max(charH, g.offsetHeight);

    g.style.lineHeight = '1';
    charLH = g.offsetHeight;

    charOffsetY = Math.round((charH - charLH) / 2);
  })();

  // Helper for creating canvases.
  function createCanvas(attach, styles, alpha) {
    var canvas = document.createElement('canvas');
    canvas.style.position = 'fixed';
    canvas.style.top = '0px';
    canvas.style.left = '0px';

    for (var prop in styles) {
      canvas.style[prop] = styles[prop];
    }

    var ctx = canvas.getContext('2d', {'alpha': !!alpha});

    if (attach) {
      document.body.appendChild(canvas);
    }

    ctx.imageSmoothingEnabled = false;
    canvas.ctx = ctx;
    return canvas;
  }

  var main = createCanvas(true, {'zIndex': 1});
  var cursor = createCanvas(true, {'zIndex': 2, 'pointerEvents': 'none'});

  // Two scratch canvases so that at least two functions can render scratch
  // images at a time.
  var _scratchI = 0;
  var _scratch = [createCanvas(), createCanvas()];
  var buffer = createCanvas();

  // Scratch buffer for transparent renders.
  var _alphaScratchI = 0;
  var _alphaScratch = [createCanvas(false, {}, true)];

  var debugColors = [
    [rgb(0xdecf3f), rgba(0xdecf3f, 0.5)],
    [rgb(0x60bd68), rgba(0x60bd68, 0.5)],
    [rgb(0xf15854), rgba(0xf15854, 0.5)],
    [rgb(0x5da5da), rgba(0x5da5da, 0.5)],
    [rgb(0xb276b2), rgba(0xb276b2, 0.5)],
  ];
  var debugRects = [];
  var debug = null;

  cursor.width = charW;
  cursor.height = charH;

  var attachedEvents = [];
  self.addEventListener = function(name, handler, capture) {
    if (attachedEvents.indexOf(name) !== -1) {
      // Only allow adding event listeners once.
      return;
    }

    attachedEvents.push(name);
    main.addEventListener(name, handler.bind(self), capture);
  };

  function scratch(w, h) {
    _scratchI++;
    var scr = _scratch[_scratchI % _scratch.length];
    scr.width = w;
    scr.height = h;
    return scr;
  }

  function scratchA(w, h) {
    _alphaScratchI++;
    var scr = _alphaScratch[_alphaScratchI % _alphaScratch.length];
    scr.width = w;
    scr.height = h;
    return scr;
  }

  self.charSize = function() {
    return [charW, charH];
  };

  self.gridSize = function() {
    return [gridW, gridH];
  };

  self.setSize = function(w, h) {
    if (w == gridW && h == gridH) {
      return;
    }

    buffer.width = main.width = w * charW;
    buffer.height = main.height = h * charH;

    if (debug) {
      debug.width = buffer.width;
      debug.height = buffer.height;
    }

    gridW = w;
    gridH = h;
  };

  self.clear = function(id, a, fg, bg, sp) {
    repeatCache = {};
    palette = {};
    brush = null;
    self.setPalette(id, a, fg, bg, sp);
    self.setAttributes(id);
    buffer.ctx.fillStyle = brush.bg;
    buffer.ctx.fillRect(0, 0, buffer.width, buffer.height);

    if (debug) {
      debug.width = debug.width;
    }
  };

  function rgb(n) {
    return '#' + ('000000' + n.toString(16)).substr(-6);
  }

  function rgba(n, a) {
    return 'rgba(' + (n >> 16) + ',' + ((n >> 8) & 0xff) + ',' + (n & 0xff) + ',' + a + ')';
  }

  self.setPalette = function(id, a, fg, bg, sp) {
    palette[id] = {
      'attr': a,
      'fg': rgb(fg),
      'bg': rgb(bg),
      'sp': rgb(sp),
    };
  };

  self.setAttributes = function(id) {
    brush = palette[id];
  };

  function setFont(ctx, a) {
    var f = font;

    if ((a & nmux.AttrBold) === nmux.AttrBold) {
      f = 'bold ' + f;
    }

    if ((a & nmux.AttrItalic) === nmux.AttrItalic) {
      f = 'italic ' + f;
    }

    ctx.font = f;
    ctx.textAlign = 'left';
    ctx.textBaseline = 'top';
  }

  function renderUndercurl(ctx, x, y, w, fg, bg, sp) {
    ctx.save();
    ctx.translate(0, -1);
    ctx.fillStyle = sp || brush.sp;

    for (var i = 0; i < w; i++) {
      // This is usually `sin(theta) * amplitude`, but we're only stepping by
      // 1px.
      ctx.fillRect(x + i, y + Math.sin(gX + i), 1, 1);
    }

    ctx.restore();
  }

  function renderUnderline(ctx, x, y, w, fg, bg, sp) {
    ctx.save();
    ctx.translate(0, charLH + 1);
    ctx.fillStyle = fg || brush.fg;
    ctx.fillRect(x, y, w, 1);
    ctx.restore();
  }

  function renderAttrs(ctx, x, y, w, a, fg, bg, sp) {
    a = a || brush.attr;

    if ((a & nmux.AttrUnderline) === nmux.AttrUnderline) {
      renderUnderline(ctx, x, y, w, fg, bg, sp);
    }

    if ((a & nmux.AttrUndercurl) === nmux.AttrUndercurl) {
      renderUndercurl(ctx, x, y, w, fg, bg, sp);
    }
  }

  function renderChar(ctx, x, y, c, a, fg, bg, sp) {
    ctx.save();
    ctx.translate(x, y);

    ctx.fillStyle = bg;
    ctx.fillRect(0, 0, charW * c.wcwidth(), charH);

    ctx.translate(0, charOffsetY - 2);
    ctx.fillStyle = fg;
    setFont(ctx, a);

    ctx.fillText(c, 0, 0);

    ctx.restore();
  }

  self.renderRepeatedText = function(c, index, len) {
    var x = (index % gridW) * charW;
    var y = Math.floor(index / gridW) * charH;
    var k = c + brush.fg + brush.bg + brush.sp;
    var pat = repeatCache[k];
    var cw = charW * c.wcwidth();
    var w = len * cw;

    gX = x;

    if (!pat) {
      var scr = scratch(cw, charH);
      renderChar(scr.ctx, 0, 0, c, brush.attr, brush.fg, brush.bg, brush.sp);
      debugRect(x, y, cw, charH, 0);
      pat = scr.ctx.createPattern(scr, 'repeat');
      repeatCache[k] = pat;
    }

    debugRect(x, y, w, charH, 1);
    buffer.ctx.fillStyle = pat;
    buffer.ctx.fillRect(x, y, w, charH);

    renderAttrs(buffer.ctx, x, y, w);
  };

  self.renderText = function(index, str) {
    var x = (index % gridW) * charW;
    var y = Math.floor(index / gridW) * charH;
    var w = str.length * charW;

    debugRect(x, y, w, charH, 2);

    gX = x;

    // Render text in the scratch canvas to keep unusually tall glyphs from
    // bleeding into other lines.
    var scr = scratch(w, charH);

    scr.ctx.save();
    scr.ctx.fillStyle = brush.bg;
    scr.ctx.fillRect(0, 0, w, charH);

    setFont(scr.ctx, brush.attr);
    scr.ctx.translate(0, charOffsetY - 2);
    scr.ctx.fillStyle = brush.fg;

    for (var c, i = 0, l = str.length; i < l; i++) {
      c = str.substr(i, 1);
      if (c != ' ') {
        scr.ctx.fillText(c, i * charW, 0);
      }
    }

    scr.ctx.restore();

    renderAttrs(scr.ctx, 0, 0, w);

    buffer.ctx.drawImage(scr, x, y);
  };

  var cursorDelay = 1000;

  function blinkCursor(t, counter) {
    if (counter % 2 == 0) {
      cursor.style.visibility = 'visible';
    } else {
      cursor.style.visibility = 'hidden';
    }
  };

  self.hideCursor = function() {
    nmux.animate.remove('cursor');
    cursor.style.visibility = 'hidden';
  };

  self.showCursor = function() {
    cursor.style.visibility = 'visible';

    switch (self.mouse.event) {
      case 'mousemove':
      case 'mousedown':
        return;
    }

    nmux.animate.add('cursor', blinkCursor, 500, cursorDelay);
  };

  self.setCursor = function(mode, x, y, id, c) {
    var b = palette[id];
    if (!b) {
      return;
    }

    c = String.fromCharCode(c);
    var cw = charW * c.wcwidth();
    gX = x * charW;
    cursor.width = cw;
    cursor.style.left = gX + 'px';
    cursor.style.top = (y * charH) + 'px';

    var a = b.attr;
    var fg = b.fg;
    var bg = b.bg;
    var sp = b.sp;

    cursorX = x;
    cursorY = y;

    if ((mode & nmux.ModeNormal) === nmux.ModeNormal) {
      var t = fg;
      fg = bg;
      bg = t;
      cursorDelay = 1000;
    } else {
      cursorDelay = 500;
    }

    renderChar(cursor.ctx, 0, 0, c, a, fg, bg, sp);
    renderAttrs(cursor.ctx, 0, 0, cw, a, fg, bg, sp);

    if ((mode & nmux.ModeInsert) === nmux.ModeInsert) {
      cursor.ctx.fillStyle = fg;
      cursor.ctx.fillRect(0, 0, 1, charH);
    } else if ((mode & nmux.ModeReplace) === nmux.ModeReplace) {
      cursor.ctx.fillStyle = fg;
      cursor.ctx.fillRect(0, charH - 3, cw, 3);
    }
  };

  self.scroll = function(bg, delta, x1, y1, x2, y2) {
    self.hideCursor();

    var w = ((x2 - x1) + 1) * charW;
    var h = ((y2 - y1) + 1) * charH;
    delta *= charH;
    x1 *= charW;
    y1 *= charH;

    var scr = scratch(w, h);

    scr.ctx.fillStyle = rgb(bg);
    scr.ctx.fillRect(0, 0, w, h);
    scr.ctx.drawImage(buffer, x1, y1, w, h, 0, -delta, w, h);

    if (debug) {
      debugDraw();
      var ascr = scratchA(w, h);
      ascr.ctx.drawImage(debug, x1, y1, w, h, 0, -delta, w, h);
      debug.ctx.clearRect(x1, y1, w, h);
      debug.ctx.drawImage(ascr, x1, y1);
    }

    buffer.ctx.fillStyle = rgb(bg);
    buffer.ctx.fillRect(x1, y1, w, h);
    buffer.ctx.drawImage(scr, x1, y1);
  };

  var debugTimer = 0;

  self.setDebug = function(d) {
    if (d) {
      if (debug) {
        return;
      }

      debug = createCanvas(true, {'zIndex': 3, 'pointerEvents': 'none'}, true);
      debug.width = buffer.width;
      debug.height = buffer.height;

      var msg = 'Debug Enabled';
      debug.ctx.save();
      setFont(debug.ctx, 0);

      var w = msg.length * charW;
      debug.ctx.translate(debug.width - w - (charW * 4), charH);
      debug.ctx.fillStyle = '#fff';
      debug.ctx.fillRect(0, 0, w + (charW * 4), charH * 3);

      debug.ctx.fillStyle = '#000';
      debug.ctx.fillText(msg, charW * 2, charH);
      debug.ctx.restore();
    } else if (debug) {
      debugRects = [];
      debug.parentNode.removeChild(debug);
      debug = null;
    }
  };

  function debugDraw() {
    clearInterval(debugTimer);

    if (debugRects.length === 0) {
      return;
    }

    var r, c, i = 0, ctx = debug.ctx;

    ctx.save();
    ctx.globalAlpha = 0.8;
    ctx.globalCompositeOperation = 'destination-in';
    ctx.drawImage(debug, 0, 0);
    ctx.restore();

    while (debugRects.length) {
      r = debugRects.pop();
      c = debugColors[r[4]];
      ctx.save();
      ctx.lineWidth = 1;
      ctx.strokeStyle = c[0];
      ctx.fillStyle = c[1];
      ctx.translate(r[0], r[1]);
      ctx.fillRect(1, 1, r[2] - 2, r[3] - 2);
      ctx.translate(0.5, 0.5);
      ctx.strokeRect(1, 1, r[2] - 2, r[3] - 2);
      ctx.restore();
      i++;
    }
  }

  function debugRect(x, y, w, h, color) {
    if (!debug) {
      return;
    }
    debugRects.push([x, y, w, h, color]);
  }

  self.flush = function() {
    main.ctx.drawImage(buffer, 0, 0);

    if (debug) {
      clearInterval(debugTimer);
      debugTimer = setTimeout(debugDraw, 10);
    }
  };
}
