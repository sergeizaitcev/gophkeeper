package hashio_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/pkg/hashio"
)

var (
	testdata = []byte("test data")
	checksum = "916f0027a575074ce72a331777c3478d6513f786a591bd892da1a577bf2335f9"
)

func TestHashReader(t *testing.T) {
	hr := hashio.NewHashReader(bytes.NewReader(testdata))
	_, _ = io.Copy(io.Discard, hr)
	require.Equal(t, checksum, hr.Checksum())
}

func TestHashWriter(t *testing.T) {
	hw := hashio.NewHashWriter(io.Discard)
	_, _ = io.Copy(hw, bytes.NewReader(testdata))
	require.Equal(t, checksum, hw.Checksum())
}
