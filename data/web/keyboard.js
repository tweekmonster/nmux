'use strict';

function nmuxMouseKey(e) {
  var prefix = '';
  var suffix = 'Mouse';

  var charSize = this.charSize();
  var x = Math.floor(e.x / charSize[0]);
  var y = Math.floor(e.y / charSize[1]);

  if (event.type == this.mouse.event && x == this.mouse.x && y == this.mouse.y) {
    return;
  }

  this.mouse.event = event.type;
  this.mouse.x = x;
  this.mouse.y = y;

  switch (e.type) {
    case 'mousemove':
      suffix = 'Drag';
      break;
    case 'mouseup':
      suffix = 'Release';
      this.mouse.ox = -1;
      this.mouse.oy = -1;
      this.showCursor();
      break;
    case 'mousedown':
      this.mouse.ox = x;
      this.mouse.oy = y;
  }

  switch (e.button) {
    case 0:
      prefix = 'Left';
      break;

    case 1:
      prefix = 'Middle';
      break;

    case 2:
      prefix = 'Right';
      break;
  }

  return '<' + prefix + suffix + '><' + x + ',' + y + '>';
}

function nmuxMouseWheel(e) {
  var key = 'ScrollWheelDown';

  if (e.deltaY < 0) {
    key = 'ScrollWheelUp';
  }

  if (e.shiftKey) {
    key = 'S-' + key;
  }

  if (e.ctrlKey) {
    key = 'C-' + key;
  }

  if (e.altKey) {
    key = 'A-' + key;
  }

  if (e.metaKey) {
    key = 'D-' + key;
  }

  return '<' + key + '>';
}

function nmuxKey(e) {
  var key = e.key;
  var shift = e.shiftKey;

  if (e.code.substr(0, 3) == 'Key') {
    key = e.code.substr(3);
    if (shift) {
      key = key.toUpperCase();
    } else {
      key = key.toLowerCase();
    }
  } else if (e.code.substr(0, 6) == 'Numpad') {
    key = 'key' + e.code.substr(6);
    switch (key) {
      case 'kDecimal':
        key = 'kPoint';
        break;
      case 'kAdd':
        key = 'kPlus';
        break;
      case 'kSubtract':
        key = 'kMinus';
        break;
    }
  } else if (e.code.substr(0, 5) == 'Arrow') {
    key = e.code.substr(5);
  } else if (e.code.substr(0, 5) == 'Digit') {
    shift = false;
  }

  switch (e.key.toLowerCase()) {
    case 'ctrl':
    case 'control':
    case 'alt':
    case 'shift':
    case 'meta':
      return;
    case ' ':
      key = 'Space';
      break;
    case '<':
      key = ',';
      break;
    case '>':
      key = '.';
      break;
    case '|':
      key = 'Bar';
      shift = false;
      break;
    case '\\':
      key = 'Bslash';
      break;
    case 'escape':
      key = 'Esc';
      break;
    case 'delete':
      key = 'Del';
      break;
    case 'backspace':
      key = 'BS';
      break;
    case '`':
    case '~':
    case '-':
    case '_':
    case '=':
    case '+':
    case '[':
    case '{':
    case '}':
    case ']':
    case ':':
    case ';':
    case '\'':
      shift = false;
      break;
  }

  if (shift && (key.length > 1 || key.toLowerCase() == key.toUpperCase())) {
    key = 'S-' + key;
  }

  if (e.ctrlKey) {
    key = 'C-' + key;
  }

  if (e.altKey) {
    key = 'A-' + key;
  }

  if (e.metaKey) {
    key = 'D-' + key;
  }

  if (key.length > 1) {
    key = '<' + key + '>';
  }

  return key
}
