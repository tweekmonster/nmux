# nmux

`nmux` is a multiplexer for [Neovim][] processes.  It is very much a work in
progress at the moment.

It currently has a built-in HTTP server that renders a single `nvim` process in
your browser using websockets as a proof-of-concept.  Short video of `nvim`
rendered in a browser: https://youtu.be/mzfHBPHkT-E

## Install

```
$ go install github.com/tweekmonster/nmux/cmd/nmux
```

## Usage

```
$ nmux --addr localhost:9999
```

Then point your browser to [http://localhost:9999/](http://localhost:9999/)


## Goals

- A server that manages multiple `nvim` processes.
  - Allow clients to connect over TCP.
- Native cross-platform client programs.
  - Each `nvim` instance can be a tab or a split view.
  - UI is always consistent.  No platform-specific GUI elements, except for the
    title bar.
  - Image replacements for glyphs (in-editor icons).
  - Basic OS integration (clipboard, notifications, open URLs, etc.)
- "Simplified" configuration.
  - Only basics need to be configured for client programs.
  - No need to configure a terminal emulator or tmux.  `nvim` can already be
    configured and scripted to no end.

The ultimate goal is to make the terminal emulator an obsolete program in my
daily work.  `nvim` has reliable terminal emulation built-in through
[libvterm][].

This will make it possible to turn a Docker container or Virtual Machine
(possibly even WSL) into your "IDE".  You could take a snapshot of your
workspaces and resume where you left off after a reboot.


[Neovim]: https://github.com/neovim/neovim
[libvterm]: https://github.com/neovim/libvterm
