package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gopherjs/gopherjs/js"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/tlc"
)

func main() {
	js.Global.Set("Wharf", map[string]interface{}{
		"diff": Diff,
	})
}

// Diff lets one create patches
func Diff(signatureBytes *js.Object, jsContainer *js.Object, callbacks *js.Object) {
	go func() {
		startTime := time.Now()

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
		container.Size = offset

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

		patchBuf := new(bytes.Buffer)
		signatureBuf := new(bytes.Buffer)

		consumer := pwr.StateConsumer{}

		if callbacks.Bool() {
			if onMessage := callbacks.Get("onMessage"); onMessage != nil {
				consumer.OnMessage = func(level string, msg string) {
					onMessage.Invoke(level, msg)
				}
			}

			if onProgress := callbacks.Get("onProgress"); onProgress != nil {
				consumer.OnProgress = func(perc float64) {
					onProgress.Invoke(perc)
				}
			}
		}

		dctx := &pwr.DiffContext{
			TargetContainer: targetContainer,
			TargetSignature: targetSignature,

			SourceContainer: container,
			FilePool:        hp,

			Compression: &pwr.CompressionSettings{
				Algorithm: pwr.CompressionAlgorithm_GZIP,
				Quality:   1,
			},

			Consumer: &consumer,
		}

		err = dctx.WritePatch(patchBuf, signatureBuf)
		if err != nil {
			panic(err)
		}

		consumer.Infof("%s patch", humanize.Bytes(uint64(patchBuf.Len())))
		consumer.Infof("%s signature", humanize.Bytes(uint64(signatureBuf.Len())))

		prettySize := humanize.Bytes(uint64(targetContainer.Size))
		perSecond := humanize.Bytes(uint64(float64(targetContainer.Size) / time.Since(startTime).Seconds()))
		consumer.Infof("%s (%s) @ %s/s\n", prettySize, targetContainer.Stats(), perSecond)
	}()
}
