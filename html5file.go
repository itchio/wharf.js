package main

import (
	"io"
	"os"

	"github.com/flimzy/jsblob"
)

type html5File struct {
	blob   jsblob.Blob
	offset int64
	size   int64
	closed bool
}

var _ io.ReadCloser = (*html5File)(nil)
var _ io.Seeker = (*html5File)(nil)
var _ io.ReaderAt = (*html5File)(nil)

func (h5 *html5File) Read(buf []byte) (n int, err error) {
	start := h5.offset
	end := start + int64(len(buf))
	if end > h5.size {
		err = io.EOF
		end = h5.size - 1
	}

	bs := h5.blob.Slice(int(start), int(end), "").Bytes()
	copy(buf, bs)

	n = int(end - start)
	if n > 0 {
		err = nil
	}

	h5.offset += (end - start)
	return
}

func (h5 *html5File) ReadAt(buf []byte, off int64) (n int, err error) {
	_, err = h5.Seek(off, os.SEEK_SET)
	if err != nil {
		return
	}

	n, err = h5.Read(buf)
	return
}

func (h5 *html5File) Seek(whence int64, mode int) (n int64, err error) {
	switch mode {
	case os.SEEK_CUR:
		h5.offset += whence
	case os.SEEK_SET:
		h5.offset = whence
	case os.SEEK_END:
		h5.offset = h5.size - whence
	}

	if h5.offset < 0 {
		h5.offset = 0
	}

	if h5.offset >= h5.size {
		h5.offset = h5.size
	}

	n = h5.offset

	return
}

func (h5 *html5File) Close() (err error) {
	h5.closed = true
	return
}
