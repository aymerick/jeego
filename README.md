jeego
=====

House monitoring with Jeenode/TinyTX sensors and Go lang.

![Jeego Logo](https://github.com/aymerick/jeego/blob/master/jeego.png?raw=true "Jeego")


Install
=======

```bash
$ go get github.com/aymerick/jeego/jeego
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
$ cd ~/Dev/go/src/github.com/aymerick/jeego/jeego
$ go build
$ ./jeego
```


Test
====

```bash
$ cd ~/Dev/go/src/github.com/aymerick/jeego
go test ./...
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
  "log_level": "info",
  "log_file": "stdout"
}
```


Nodes kinds
===========

See [jeego-devices](https://github.com/aymerick/jeego-devices) repo.


Todo
====

- Web API:
  * List all nodes
  * Get/Update/Delete a node
  * Disable/enable a node sensor
  * Set node sensor shift correction
- Web client:
  * Use Web API
  * Display graphs, updated with websockets
- Auto-shift-correction mode: select a list of nodes, puts them in the same room during 24h => automatically set sensors shift corrections
- 'Room' concept (eg: you can have several temp sensors in the same room)
- New sensors:
  * Current usage
  * Leak detector
  * Smoke detector
  * CO2 detector
- Actuators:
  * Light switch
  * Squeezebox
- Scenario engine (with LUA ?). Examples:
  * Start radio on squeezebox in kitchen on first motion after 7:00 AM
  * Switch on low dimmed toilets light on motion during the night
- Daemonize jeego
- Munin plugin
- Debian package
