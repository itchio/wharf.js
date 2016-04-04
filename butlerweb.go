package main

import (
	"io"
	"log"
	"os"

	"github.com/flimzy/jsblob"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("wharf", map[string]interface{}{
		"Diff": Diff,
	})
	//
	// log.Println("GOOS: ", runtime.GOOS)
	//
	// TargetContainer := &tlc.Container{
	// 	Files: []*tlc.File{
	// 		{Path: "hello.txt", Size: 1024},
	// 	},
	// }
	// SourceContainer := &tlc.Container{
	// 	Files: []*tlc.File{
	// 		{Path: "hello.txt", Size: 1024},
	// 	},
	// }
	//
	// patchBuf := new(bytes.Buffer)
	// signatureBuf := new(bytes.Buffer)
	//
	// dctx := &pwr.DiffContext{
	// 	Compression: &pwr.CompressionSettings{
	// 		Algorithm: pwr.CompressionAlgorithm_UNCOMPRESSED,
	// 		Quality:   0,
	// 	},
	// 	Consumer: &pwr.StateConsumer{
	// 		OnMessage: func(level string, msg string) {
	// 			log.Printf("[%s] %s\n", level, msg)
	// 		},
	// 	},
	//
	// 	TargetContainer: TargetContainer,
	// 	SourceContainer: SourceContainer,
	// }
	//
	// log.Println("Writing patch...")
	//
	// err := dctx.WritePatch(patchBuf, signatureBuf)
	// if err != nil {
	// 	log.Println("Patching error: ", err.Error())
	// }
	//
	// log.Printf("Patch: %v", patchBuf.Bytes())
	// log.Printf("Signature: %v", signatureBuf.Bytes())
	//
	// log.Println("Patch generated!")
}

type html5File struct {
	blob   jsblob.Blob
	offset int64
	size   int64
	closed bool
}

var _ io.ReadCloser = (*html5File)(nil)
var _ io.Seeker = (*html5File)(nil)

// Diff runs some tests right now
func Diff(file *js.Object) {
	go func() {
		log.Println("In diff!")
		r := &html5File{
			blob:   jsblob.Blob{*file},
			offset: 0,
			size:   file.Get("size").Int64(),
			closed: false,
		}

		log.Println("Copying..")
		io.Copy(os.Stdout, r)
		log.Println("Done copying!")
	}()
}

func (h5 *html5File) Read(buf []byte) (n int, err error) {
	log.Printf("Should read %d bytes\n", len(buf))

	start := h5.offset
	end := start + int64(len(buf))
	if end > h5.size {
		err = io.EOF
		end = h5.size - 1
	}

	log.Printf("Reading %d bytes\n", end-start)

	bs := h5.blob.Slice(int(start), int(end), "").Bytes()
	copy(buf, bs)

	n = int(end - start)
	if n > 0 {
		err = nil
	}

	h5.offset += (end - start)
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
