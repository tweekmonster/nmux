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
  var attrs = 0;
  var colors = [0, 0, 0];
  var fgColor = '#fff';
  var bgColor = '#000';
  var spColor = '#000';

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
  function createCanvas(attach, styles) {
    var canvas = document.createElement('canvas');
    canvas.style.position = 'fixed';
    canvas.style.top = '0px';
    canvas.style.left = '0px';

    for (var prop in styles) {
      canvas.style[prop] = styles[prop];
    }

    var ctx = canvas.getContext('2d', {'alpha': false});

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

    gridW = w;
    gridH = h;
  };

  self.clear = function() {
    buffer.ctx.fillStyle = bgColor;
    buffer.ctx.fillRect(0, 0, buffer.width, buffer.height);
  };

  function rgb(n) {
    return '#' + ('000000' + n.toString(16)).substr(-6);
  }

  self.setAttributes = function(a, fg, bg, sp) {
    attrs = a;
    colors = [fg, bg, sp];
    fgColor = rgb(fg);
    bgColor = rgb(bg);
    spColor = rgb(sp);
  }

  function setFont(ctx, attrs) {
    var f = font;

    if ((attrs & nmux.AttrBold) === nmux.AttrBold) {
      f = 'bold ' + f;
    }

    if ((attrs & nmux.AttrItalic) === nmux.AttrItalic) {
      f = 'italic ' + f;
    }

    ctx.font = f;
    ctx.textAlign = 'left';
    ctx.textBaseline = 'top';
  }

  function renderUndercurl(ctx, x, y, w, fg, bg, sp) {
    ctx.save();
    ctx.translate(0, -1);
    ctx.fillStyle = sp || spColor;

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
    ctx.fillStyle = fg || fgColor;
    ctx.fillRect(x, y, w, 1);
    ctx.restore();
  }

  function renderAttrs(ctx, x, y, w, a, fg, bg, sp) {
    a = a || attrs;

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
    ctx.fillRect(0, 0, charW, charH);

    ctx.translate(0, charOffsetY - 2);
    ctx.fillStyle = fg;
    setFont(ctx, a);

    ctx.fillText(c, 0, 0);

    ctx.restore();
  }

  var repeatCache = {};

  self.renderRepeatedText = function(c, index, len) {
    var x = (index % gridW) * charW;
    var y = Math.floor(index / gridW) * charH;
    var k = fgColor + bgColor + spColor;
    var pat = repeatCache[k];

    gX = x;

    if (c != ' ' || !pat) {
      var scr = scratch(charW, charH);
      renderChar(scr.ctx, 0, 0, c, attrs, fgColor, bgColor, spColor);

      pat = scr.ctx.createPattern(scr, 'repeat');

      if (c == ' ') {
        // Cache repeated spaces.
        repeatCache[k] = pat;
      }
    }

    var w = len * charW;

    buffer.ctx.fillStyle = pat;
    buffer.ctx.fillRect(x, y, w, charH);

    renderAttrs(buffer.ctx, x, y, w);
  };

  self.renderText = function(str, index) {
    var x = (index % gridW) * charW;
    var y = Math.floor(index / gridW) * charH;
    var w = str.length * charW;

    gX = x;

    // Render text in the scratch canvas to keep unusually tall glyphs from
    // bleeding into other lines.
    var scr = scratch(w, charH);

    scr.ctx.save();
    scr.ctx.fillStyle = bgColor;
    scr.ctx.fillRect(0, 0, w, charH);

    setFont(scr.ctx, attrs);
    scr.ctx.translate(0, charOffsetY - 2);
    scr.ctx.fillStyle = fgColor;

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

  self.setCursor = function(mode, x, y, c, a, fg, bg, sp) {
    cursor.style.left = (x * charW) + 'px';
    cursor.style.top = (y * charH) + 'px';

    gX = x * charW;

    c = String.fromCharCode(c);
    fg = rgb(fg);
    bg = rgb(bg);
    sp = rgb(sp);
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
    renderAttrs(cursor.ctx, 0, 0, charW, a, fg, bg, sp);

    if ((mode & nmux.ModeInsert) === nmux.ModeInsert) {
      cursor.ctx.fillStyle = fg;
      cursor.ctx.fillRect(0, 0, 1, charH);
    } else if ((mode & nmux.ModeReplace) === nmux.ModeReplace) {
      cursor.ctx.fillStyle = fg;
      cursor.ctx.fillRect(0, charH - 3, charW, 3);
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

    buffer.ctx.fillStyle = rgb(bg);
    buffer.ctx.fillRect(x1, y1, w, h);
    buffer.ctx.drawImage(scr, x1, y1);
  };

  self.flush = function() {
    main.ctx.drawImage(buffer, 0, 0);
  };
}
