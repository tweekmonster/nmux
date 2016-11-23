'use strict';

nmux.socket = function(url) {
  var self = this;
  var sock = null;
  var retryTimer = 0;
  var retryAttempts = 0;
  var events = document.createDocumentFragment();

  self.addEventListener = events.addEventListener.bind(events);
  self.removeEventListener = events.removeEventListener.bind(events);
  self.dispatchEvent = events.dispatchEvent.bind(events);

  function forwardEvent(e) {
    self.dispatchEvent(new e.constructor(e.type, e));
  }

  function sockClose(e) {
    if (e.wasClean) {
      forwardEvent(e);
      return;
    }

    if (retryAttempts >= 100) {
      retryAttempts = 0;
      console.error('Gave up trying to reconnect.');
      return;
    }

    // Slowly back off from retrying connections.
    // log10((0,100)) is ~2.5 minutes worth of delays + the timeout for each
    // attempt.
    retryAttempts++;
    retryConnection(1000 * Math.log10(retryAttempts));
  }

  function sockOpen(e) {
    clearInterval(retryTimer);
    console.info(retryAttempts === 0 ? 'Connected' : 'Reconnected');
    retryAttempts = 0;
    forwardEvent(e);
  }

  function retryConnection(delay) {
    clearInterval(retryTimer);
    retryTimer = setTimeout(connect, delay);
  }

  function connect() {
    sock = new WebSocket(url);
    sock.binaryType = 'arraybuffer';
    sock.addEventListener('open', sockOpen);
    sock.addEventListener('close', sockClose);
    sock.addEventListener('message', forwardEvent);

    // The error event isn't very useful.
    // See: https://www.w3.org/TR/websockets/#concept-websocket-close-fail
    sock.addEventListener('error', forwardEvent);
  }

  self.connect = function() {
    if (sock.readyState === WebSocket.CLOSED && retryAttempts === 0) {
      connect();
      return true;
    }
    return false;
  };

  self.send = function(data) {
    if (sock.readyState !== WebSocket.OPEN) {
      var err = '';

      switch(sock.readyState) {
        case WebSocket.CLOSED:
          err = 'Closed';
          break;
        case WebSocket.CLOSING:
          err = 'Closing';
          break;
        case WebSocket.CONNECTING:
          err = 'Connecting';
          break;
      }

      console.warn('Message not sent.  Connection state:', err,
                   '(' + sock.readyState + ')');
      return false;
    }

    sock.send(data);
    return true;
  };

  connect();
};

nmux.socket.prototype = EventTarget
