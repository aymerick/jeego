jeego
=====

House monitoring with Jeenode/TinyTX sensors and Go lang.


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
  "log_level": "info",
  "log_file": "stdout"
}
```


Nodes kinds
===========

See [jeego-devices](https://github.com/aymerick/jeego-devices) repo.
