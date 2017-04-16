# Fmaze [![Travis CI](https://travis-ci.org/telendt/fmaze.svg?branch=master)](https://travis-ci.org/telendt/fmaze) [![codecov](https://codecov.io/gh/telendt/fmaze/branch/master/graph/badge.svg)](https://codecov.io/gh/telendt/fmaze) [![Go Report Card](https://goreportcard.com/badge/github.com/telendt/fmaze)](https://goreportcard.com/report/github.com/telendt/fmaze)

Fmaze is a simple TCP server that reads events from an *event source*
and forwards them when appropriate to *user clients*.

It's a solution to a programming challenge that some company sends to
its candidates.

## Usage

    $ ./fmaze -h
    Usage of ./fmaze:
      -auth-timeout duration
            Client authentication timeout (default 1s)
      -clients-listen string
            User clients listen address (default ":9099")
      -event-source-listen string
            Event source listen address (default ":9090")
      -events-capacity int
            Capacity of unordered events store (default 100000)
      -flush-interval duration
            Write flush interval (default 10s)
      -msg-backlog int
            Client message backlog (default 10)
      -no-backpressure
            Disable client write backpressure
      -no-reset
            Don't reset internal state when event source disconnects
      -read-buffer int
            Read buffer size in bytes (default 4096)
      -start-sequence int
            Sequence start number (default 1)
      -use-writev
            Try to use writev instead of write syscall
      -write-buffer int
            Write buffer size in bytes (default 4096)
