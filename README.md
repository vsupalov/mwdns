Installation
============

TODO :) with `go get`. Maybe even explain installing/setting up go?

Node dependencies for minification:
```
sudo npm install -g less uglify-js
```

Running
=======

First, compile the server by simply running `go build`.

By default, the server listens on `localhost:8080`, in order to specify
a different port to listen to, but also to listen to an interface which is
opened to the world, set the `-addr` to your hostname or public IP:

```bash
$ ./mwdns -addr jupiler:9000
```

Positions and coordinates
=========================
Coordinate-system is the javascript one, (0,0) being top left, x increasing towards the right, y increasing towards the bottom.

Client-side (js), all positions are stored in pixels.
In the messages aswell as on the server (go), all positions are relative in % of game board size, i.e. between 0.0 and 1.0 if on the board.
