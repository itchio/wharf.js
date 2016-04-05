package main

import (
	"compress/gzip"
	"io"

	"github.com/dsnet/compress/brotli"
	"github.com/itchio/wharf/pwr"
)

type gzipCompressor struct{}

func (gc *gzipCompressor) Apply(writer io.Writer, quality int32) (io.Writer, error) {
	gw := gzip.NewWriter(writer)
	return gw, nil
}

type brotliDecompressor struct{}

func (bc *brotliDecompressor) Apply(reader io.Reader) (io.Reader, error) {
	br := brotli.NewReader(reader)
	return br, nil
}

func init() {
	pwr.RegisterCompressor(pwr.CompressionAlgorithm_GZIP, &gzipCompressor{})
	pwr.RegisterDecompressor(pwr.CompressionAlgorithm_BROTLI, &brotliDecompressor{})
}
