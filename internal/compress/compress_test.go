package compress

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var src = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Maecenas sagittis sapien non risus sollicitudin malesuada."
var dst = []byte{0x78, 0xda, 0x14, 0xc6, 0xd1, 0x89, 0x3, 0x31, 0xc, 0x4, 0xd0, 0x56, 0xa6, 0x80, 0x63, 0x2b, 0xb9, 0x14, 0x21, 0x6c, 0xb1, 0xc, 0xd8, 0x92, 0xf1, 0x48, 0xfd, 0x87, 0xfc, 0xbd, 0xff, 0xbc, 0xbe, 0xc1, 0xa3, 0xde, 0x98, 0xb9, 0xf2, 0x42, 0x2c, 0xd8, 0xf6, 0xfa, 0xc3, 0xc8, 0x90, 0x8f, 0xf2, 0xea, 0xb, 0x9b, 0x3c, 0xd4, 0x60, 0xbc, 0xf0, 0xc5, 0x7a, 0xf0, 0x31, 0x1f, 0x1e, 0x26, 0xc8, 0x5e, 0x56, 0xf1, 0x87, 0x43, 0xf, 0x44, 0x6, 0x2e, 0xd5, 0x82, 0x72, 0x2d, 0xe, 0x56, 0x4f, 0x6, 0xb6, 0x2d, 0x57, 0xdb, 0xb4, 0xe7, 0x1b, 0x0, 0x0, 0xff, 0xff, 0xbf, 0x8d, 0x2b, 0x4b}

var src1 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
var dst1 = []byte{0x78, 0xda, 0x72, 0x1c, 0xe4, 0x0, 0x10, 0x0, 0x0, 0xff, 0xff, 0xc7, 0xa4, 0x28, 0xa1}

func Test_Compress(t *testing.T) {
	r := bytes.NewBufferString(src)
	compressedBytes, _ := ioutil.ReadAll(ZlibCompress(r))
	assert.Equal(t, 0, bytes.Compare(compressedBytes, dst))

	r = bytes.NewBufferString(src1)
	compressedBytes, _ = ioutil.ReadAll(ZlibCompress(r))
	assert.Equal(t, 0, bytes.Compare(compressedBytes, dst1))
}
