# gomost

[![Go Report Card](https://goreportcard.com/badge/github.com/landonia/goprox)](https://goreportcard.com/report/github.com/landonia/goprox)

A simple proxy allowing you to reverse proxy requests to other paths.

## Overview

goprox (GO Proxy) allows you to simply reverse proxy requests. It is useful to use locally when you do not wish o(or need) apache or nginx.

## Installation

With a healthy Go Language installed, simply run `go get github.com/landonia/goprox`

### Static Host Sites

Then to run the proxy you can simply execute `goprox` within the root directory
containing the static resources.

By default, the current directory where the program is executed will be used
as the root static folder. If you wish to change this, you can provide a configuration
file such as:

```
  host: :8080
  StaticDir: /the/path/to/the/root/dir
```

then run `goprox -c=myconf.yaml`

### Application Proxy

If you wish to proxy requests to another application you need to provide a YAML configuration file that provides the proxy host mappings.

```
  host: :8080
  proxies:
    -
      proxy: /yourapi/something/else
      to: http://localhost:8090/api/else
    -
      proxy: /myapi/something/else
      to: http://localhost:8091/api/something
```

then run `goprox -c=myconf.yaml`

### Config Options

There are multiple other configuration properties than can be provided to the program.

```
  host: :80 // The local address - Set to ':80' when in production
  loglevel: fatal|error|warn|info|debug|trace // info by default
  static: /the/path/to/the/root/dir // The location of the static resources
  proxies:
    -
      proxy: /yourapi/something/else
      to: http://localhost:8090/api/else
    -
      proxy: /myapi/something/else
      to: http://localhost:8091/api/something
```

## About

goproxwas written by [Landon Wainwright](http://www.landotube.com) | [GitHub](https://github.com/landonia).

Follow me on [Twitter @landotube](http://www.twitter.com/landotube)! Although I don't really tweet much tbh.
