jeego
=====

House monitoring with Jeenode sensors and Go lang


Install
=======

```bash
$ go get github.com/aymerick/jeego
```


Update
======

```bash
$ go get -u github.com/aymerick/jeego
```


Run
===

```bash
$ jeego
```

Default conf file is `~/.jeego.json` but you can change location with:

```bash
$ JEEGO_CONFIG='<path_to_conf_file>' jeego
```


Dev
===

```bash
$ ln -s ~/Dev/go/src/github.com/aymerick/jeego/.jeego.json ~/
```

```bash
$ cd ~/Dev/go/src/github.com/aymerick/jeego
$ go build
$ ./jeego
```


Conf file
=========

Example:

```json
{
  "serial_port": "/dev/tty.usbserial-A1014IM4",
  "serial_baud": 57600,
  "domoticz_host": "127.0.0.1",
  "domoticz_port": 8080,
  "debug": true,
  "nodes": [
    {
      "id": 2,
      "kind": "roomNode",
      "name": "Bureau Aymerick",
      "domoticz_idx": "2"
    },
    {
      "id": 3,
      "kind": "thlNode",
      "name": "thlNode test",
      "domoticz_idx": "3"
    }
  ]
}
```


Nodes kinds
===========

- roomNode: Jeelabs official Room Board (http://jeelabs.com/products/room-board)
- thlNode: "Temperature Humidity Light" with DHT22 and LDR
