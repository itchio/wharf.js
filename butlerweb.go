package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/dustin/go-humanize"
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

// Diff runs some tests right now
func Diff(file *js.Object) {
	go func() {
		log.Println("In diff!")
		h5 := &html5File{
			blob:   jsblob.Blob{*file},
			offset: 0,
			size:   file.Get("size").Int64(),
			closed: false,
		}

		t1 := time.Now()
		filebytes := h5.blob.Bytes()

		l := time.Since(t1)
		log.Printf("Took %s to retrieve whole file", l.String())

		r := bytes.NewReader(filebytes)
		fullyReadZip(r, int64(len(filebytes)))

		// fullyReadZip(h5, h5.size)
	}()
}

func fullyReadZip(r io.ReaderAt, size int64) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		panic(err)
	}

	var total int64
	t1 := time.Now()

	for _, file := range zr.File {
		log.Printf("%20s %s\n", humanize.Bytes(uint64(file.FileInfo().Size())), file.Name)
		rc, err := file.Open()
		if err != nil {
			panic(err)
		}

		readBytes, err := io.Copy(ioutil.Discard, rc)
		if err != nil {
			panic(err)
		}

		total += readBytes

		err = rc.Close()
		if err != nil {
			panic(err)
		}
	}

	len := time.Since(t1)
	fmt.Printf("Total contents size: %s (read in %s, %s / s)\n",
		humanize.Bytes(uint64(total)), len.String(),
		humanize.Bytes(uint64(float64(total)/len.Seconds())))
}
