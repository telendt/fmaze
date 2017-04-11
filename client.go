package main

import (
	"bufio"
	"io"
	"net"
	"time"
)

type flushWriter interface {
	io.Writer
	Flush() error
}

type directWriter struct {
	io.Writer
}

func (d directWriter) Flush() error {
	return nil
}

type vecWriter struct {
	w       io.Writer
	bufs    net.Buffers
	bufSize int
	n       int
}

func (v *vecWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	v.n += n
	v.bufs = append(v.bufs, p)
	if v.n >= v.bufSize {
		return n, v.Flush()
	}
	return n, nil
}

func (v *vecWriter) Flush() error {
	n, err := v.bufs.WriteTo(v.w)
	v.n -= int(n)
	return err
}

// MaxLatencyForwarder takes messages from given channel and forwards them
// back to a given writer.
type MaxLatencyForwarder struct {
	flushWriterFactory func(io.Writer) flushWriter
	latency            time.Duration
}

// Forward forwards messages from src channel into a dst writer.
// It stops forwarding on src or done channel close and on any write error.
func (m MaxLatencyForwarder) Forward(done <-chan struct{}, dst io.Writer, src <-chan []byte) {
	fw := m.flushWriterFactory(dst)
	var flushC <-chan time.Time
	if m.latency > 0 {
		t := time.NewTicker(m.latency)
		defer t.Stop()
		flushC = t.C
	} else {
		flushC = make(chan time.Time)
	}
	for {
		select {
		case msg, more := <-src:
			if !more {
				fw.Flush()
				return
			}
			if _, err := fw.Write(msg); err != nil {
				return
			}
		case <-flushC:
			fw.Flush()
		case <-done:
			fw.Flush()
			return
		}
	}
}

// NewMaxLatencyForwarder returns a new MaxLatencyWriter.
func NewMaxLatencyForwarder(bufSize int, latency time.Duration, useWritev bool) MaxLatencyForwarder {
	var f func(io.Writer) flushWriter
	if bufSize > 0 {
		if useWritev {
			f = func(w io.Writer) flushWriter {
				return &vecWriter{
					w:       w,
					bufSize: bufSize,
				}
			}
		} else {
			f = func(w io.Writer) flushWriter {
				return bufio.NewWriterSize(w, bufSize)
			}
		}
	} else {
		f = func(w io.Writer) flushWriter {
			return directWriter{w}
		}
	}
	return MaxLatencyForwarder{f, latency}
}
