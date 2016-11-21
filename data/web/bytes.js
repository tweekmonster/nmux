'use strict';

var Reader = function(bytes) {
  // Because TypedArray is cumbersome.
  var self = this;
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
}
