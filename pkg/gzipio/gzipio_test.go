package gzipio_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sergeizaitcev/gophkeeper/pkg/gzipio"
	"github.com/sergeizaitcev/gophkeeper/pkg/randutil"
)

func TestCompress(t *testing.T) {
	want := randutil.Bytes(1 << 10)
	var buf bytes.Buffer

	_, err := buf.ReadFrom(gzipio.NewCompressingReader(bytes.NewReader(want)))
	require.NoError(t, err)
	require.NotEmpty(t, buf.Len())
	require.Less(t, buf.Len(), len(want))

	dc, err := gzipio.NewDecompressingReader(io.NopCloser(&buf))
	require.NoError(t, err)
	defer dc.Close()

	got, err := io.ReadAll(dc)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
