package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/dustin/go-humanize"
	"github.com/gopherjs/gopherjs/js"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/tlc"
)

func main() {
	js.Global.Set("wharf", map[string]interface{}{
		"Diff": Diff,
	})
}

// Diff is a wonderful work of wizardry
func Diff(signatureBytes *js.Object, jsContainer *js.Object) {
	go func() {
		// dirs := make(map[string]bool)
		container := &tlc.Container{}

		var entries = jsContainer.Get("entries")
		var numEntries = entries.Length()
		var offset int64

		for i := 0; i < numEntries; i++ {
			var entry = entries.Index(i)
			var size = entry.Get("size").Int64()
			container.Files = append(container.Files, &tlc.File{
				Path:   entry.Get("path").String(),
				Size:   size,
				Offset: offset,
				Mode:   0644,
			})
			offset += size
		}

		fmt.Println("Source container: ", container)

		hp := NewHTML5FilePool(jsContainer)

		nativeSignatureBytes, ok := js.Global.Get("Uint8Array").New(signatureBytes).Interface().([]byte)
		if !ok {
			panic(fmt.Errorf("Couldn't cast signatureBytes into []byte"))
		}

		log.Printf("Got %d native signature bytes\n", len(nativeSignatureBytes))
		signatureReader := bytes.NewReader(nativeSignatureBytes)

		targetContainer, targetSignature, err := pwr.ReadSignature(signatureReader)
		if err != nil {
			panic(err)
		}

		// fmt.Println("Target container: ", targetContainer)
		// fmt.Println("Target signature: ", targetSignature)

		patchBuf := new(bytes.Buffer)
		signatureBuf := new(bytes.Buffer)

		dctx := &pwr.DiffContext{
			TargetContainer: targetContainer,
			TargetSignature: targetSignature,

			SourceContainer: container,
			FilePool:        hp,

			Compression: &pwr.CompressionSettings{
				Algorithm: pwr.CompressionAlgorithm_UNCOMPRESSED,
				Quality:   0,
			},

			Consumer: &pwr.StateConsumer{
				OnMessage: func(level string, msg string) {
					log.Printf("[%s] %s\n", level, msg)
				},
				// OnProgress: func(perc float64) {
				// 	log.Printf("Progress %.2f", perc)
				// },
			},
		}

		err = dctx.WritePatch(patchBuf, signatureBuf)
		if err != nil {
			panic(err)
		}

		log.Println(humanize.Bytes(uint64(patchBuf.Len())), "patch")
		log.Println(humanize.Bytes(uint64(signatureBuf.Len())), "signature")

		// log.Printf("Patch: %v", patchBuf.Bytes())
		// log.Printf("Signature: %v", signatureBuf.Bytes())
	}()
}
