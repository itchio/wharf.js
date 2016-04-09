package main

import (
	"io"

	"github.com/gopherjs/gopherjs/js"
)

type html5CallbackWriter struct {
	callback *js.Object
}

var _ io.Writer = (*html5CallbackWriter)(nil)

func (w *html5CallbackWriter) Write(buf []byte) (int, error) {
	w.callback.Invoke(buf)
	return len(buf), nil
}
