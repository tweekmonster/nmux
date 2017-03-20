'use strict';

var Reader = function(bytes) {
  // Because TypedArray is cumbersome.
  var self = this;
  var lastDump = 0;
  var cursor = 0;
  var last = 0;

  bytes = new Uint8Array(bytes);

  self.remaining = function() {
    return bytes.length - cursor;
  };

  self.uint8 = function() {
    last = 1;
    return bytes[cursor++];
  };

  self.uint16 = function() {
    last = 2;
    return (bytes[cursor++] << 8) | bytes[cursor++];
  }

  self.int16 = function() {
    var u = self.uint16();
    return u > 0x8000 ? u - 0x10000 : u;
  }

  self.uint24 = function() {
    last = 3;
    return (bytes[cursor++] << 16) | (bytes[cursor++] << 8) | bytes[cursor++];
  }

  self.uint32 = function() {
    last = 4;
    return (bytes[cursor++] << 24) | (bytes[cursor++] << 16) | (bytes[cursor++] << 8) | bytes[cursor++];
  }

  self.eint32 = function() {
    var b = bytes[cursor];

    if ((b & 0x00000080) === 0) {
      cursor++;
      return b;
    }

    b = (b & 0x0000007f) | bytes[cursor + 1] << 7;
    if ((b & 0x00004000) === 0) {
      cursor += 2;
      return b;
    }

    b = (b & 0x00003fff) | bytes[cursor + 2] << 14;
    if ((b & 0x00200000) === 0) {
      cursor += 3;
      return b;
    }

    b = (b & 0x001fffff) | bytes[cursor + 3] << 21;
    if ((b & 0x10000000) === 0) {
      cursor += 4;
      return b;
    }

    b = (b & 0x0fffffff) | bytes[cursor + 4] << 28;
    cursor += 5;
    return b;
  }

  self.string = function() {
    var len = self.eint32();
    var out = '';
    while (len--) {
      out += String.fromCodePoint(self.eint32());
    }
    return out;
  }

  self.toEnd = function() {
    return bytes.subarray(cursor);
  }

  self.reverse = function() {
    self.cursor -= last;
    last = 0;
  }

  self.append = function(newBytes) {
    if (self.cursor == bytes.length) {
      last = 0;
      self.cursor = 0;
      bytes = newBytes;
      return;
    }

    var old = bytes.length;
    var b = new Uint8Array(bytes.length + newBytes.length);
    b.set(bytes, 0);
    b.set(newBytes, bytes.length);
    bytes = b;
  }

  self.dumpRange = function(i1, i2) {
    var width = 16;
    var hex = '';
    var ascii = '';
    var out = '';
    var b, i = 0, l = i2 - i1;
    var lines = [];

    while (i < l) {
      b = bytes[i + i1];
      hex += ('00' + b.toString(16)).substr(-2) + ' ';
      ascii += (b >= 32 && b < 127) ? String.fromCharCode(b) : '.';
      i++;

      if (i % width == 0) {
        lines.push(hex + ' ' + ascii);
        hex = '';
        ascii = '';
      }
    }

    if (hex.length != 0) {
      while (ascii.length < width) {
        hex += '   ';
        ascii += ' ';
      }
      lines.push(hex + ' ' + ascii);
    }

    return lines.join('\n');
  }

  self.toString = function() {
    return self.dumpRange(0, bytes.length)[0];
  }

  self.dumpLastRead = function(name) {
    var size = cursor - lastDump;
    var out = self.dumpRange(lastDump, cursor);
    lastDump = cursor;

    console.group(name + ' (' + size + ' bytes)');
    console.debug(out);
    console.groupEnd();
  }

  self.dump = function() {
    console.groupCollapsed(bytes.length + ' bytes');
    console.debug(out);
    console.groupEnd();
  }
}
