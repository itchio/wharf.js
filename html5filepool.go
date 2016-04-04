package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	gsync "sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/itchio/wharf/sync"
)

type html5FilePool struct {
	jsContainer *js.Object
	fileIndex   int64
	reader      io.ReadSeeker
}

var _ sync.FilePool = (*html5FilePool)(nil)

// NewHTML5FilePool returns a new html5 file pool..
func NewHTML5FilePool(jsContainer *js.Object) sync.FilePool {
	return &html5FilePool{
		jsContainer: jsContainer,
		fileIndex:   -1,
	}
}

func (hfp *html5FilePool) GetReader(fileIndex int64) (io.ReadSeeker, error) {
	if hfp.fileIndex == fileIndex {
		return hfp.reader, nil
	}
	hfp.fileIndex = fileIndex

	var entry = hfp.jsContainer.Get("entries").Index(int(fileIndex))
	var promise = entry.Call("read")

	var reader io.ReadSeeker
	var err error

	var wg gsync.WaitGroup
	wg.Add(1)
	promise.Call("then", func(arraybuf *js.Object) {
		bs, ok := js.Global.Get("Uint8Array").New(arraybuf).Interface().([]byte)
		if ok {
			reader = bytes.NewReader(bs)
		} else {
			log.Printf("Could not cast arraybuf to byte")
			err = fmt.Errorf("couldn't cast arraybuf to []byte")
		}
		wg.Done()
	})
	wg.Wait()

	hfp.reader = reader
	return reader, err
}

func (hfp *html5FilePool) Close() error {
	return nil
}
