# Echo Middlewares

[![Build Status](https://img.shields.io/travis/dafiti/echo-middleware/master.svg?style=flat-square)](https://travis-ci.org/dafiti/echo-middleware)
[![Coverage Status](https://img.shields.io/coveralls/dafiti/echo-middleware/master.svg?style=flat-square)](https://coveralls.io/github/dafiti/echo-middleware?branch=master)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/dafiti/echo-middleware)
[![Go Report Card](https://goreportcard.com/badge/github.com/dafiti/echo-middleware?style=flat-square)](https://goreportcard.com/report/github.com/dafiti/echo-middleware)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://github.com/dafiti/echo-middleware/blob/master/LICENSE)


Middlewares for Echo Framework

## Installation

Requires Go 1.9 or later.

```sh
go get github.com/dafiti/echo-middleware
```

## Middlewares
 - New Relic
 - LoggerWithOutput (Retrieves Logger middleware with Output)
 - Logrus (Http request logs)

## Usage Examples

```go
package main

import (
    "bytes"
    "net/http"
	mw "github.com/dafiti/echo-middleware"
    "github.com/labstack/echo"
)

func main() {
    e := echo.New()
    e.Use(mw.NewRelic("app name", "license key"))

    // Default Logger middleware
    //buf := new(bytes.Buffer)
    //e.Use(mw.LoggerWithOutput(buf))

    // Logrus HTTP request logs
    e.Use(mw.Logrus())

    e.GET("/", func(c echo.Context) error {
        txn := c.Get("newrelic-txn").(newrelic.Transaction)
        defer newrelic.StartSegment(txn, "mySegmentName").End()

        return c.String(http.StatusOK, "Hello, World!")
    })

    e.Logger.Fatal(e.Start(":1323"))
}
```

## Documentation

Read the full documentation at [https://godoc.org/github.com/dafiti/echo-middleware](https://godoc.org/github.com/dafiti/echo-middleware).

## License

This project is released under the MIT licence. See [LICENCE](https://github.com/dafiti/echo-middleware/blob/master/LICENSE) for more details.
