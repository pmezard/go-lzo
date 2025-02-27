package lzo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func testCorpus(t *testing.T, arch string, cmpfunc func([]byte) []byte) (tdata int, tcmp int) {
	t.Log("Test corpus:", arch)
	f, err := os.Open(arch)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Error(err)
		return
	}
	defer gz.Close()

	tgz := tar.NewReader(gz)
	for {
		head, err := tgz.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			return
		}

		data := make([]byte, head.Size)
		_, err = io.ReadFull(tgz, data)
		if err != nil {
			t.Error(err)
			return
		}

		cmp := cmpfunc(data)
		t.Logf("File: %-20s Size: %-10v Compressed: %-10v Factor %0.1f%%", head.Name,
			len(data), len(cmp), float32(len(data)-len(cmp))*100/float32(len(data)))

		data2, err := Decompress1X(bytes.NewReader(cmp), len(cmp), len(data))
		if err != nil {
			t.Error(err)
			continue
		}

		if !reflect.DeepEqual(data, data2) {
			t.Error("decompressed data doesn't match")
		}

		tdata += len(data)
		tcmp += len(cmp)
	}
	return
}

func testCorpora(t *testing.T, cmpfunc func([]byte) []byte) {
	archs, err := filepath.Glob("testdata/*.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	tdata, tcmp := 0, 0
	for _, arch := range archs {
		d, c := testCorpus(t, arch, cmpfunc)
		tdata += d
		tcmp += c
	}

	t.Logf("Final stats: Size: %v, Compressed: %v, Factor: %0.1f%%",
		tdata, tcmp, float32(tdata-tcmp)*100/float32(tdata))

}

func TestDecompInlen(t *testing.T) {
	data := bytes.Repeat([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 1000)
	cmp := Compress1X(data)

	for i := 1; i < 16; i++ {
		for j := -16; j < 16; j++ {
			_, err := Decompress1X(io.LimitReader(bytes.NewReader(cmp), int64(len(cmp)-i)), len(cmp)+j, 0)
			if err != io.EOF {
				t.Error("EOF expected for truncated input, found:", err)
			}
		}
	}

	for j := -16; j < 16; j++ {
		data2, err := Decompress1X(bytes.NewReader(cmp), len(cmp)+j, 0)
		if j < 0 && err != io.EOF {
			t.Error("EOF expected for truncated input, found:", err)
		}
		if j >= 0 {
			if err != nil {
				t.Error("error for normal decompression:", err, j)
			} else if !reflect.DeepEqual(data, data2) {
				t.Error("data doesn't match after decompression")
			}
		}
	}
}

func Test1(t *testing.T) {
	testCorpora(t, Compress1X)
}

func Test999(t *testing.T) {
	if !testing.Short() {
		testCorpora(t, Compress1X999)
	}
}

func BenchmarkComp(b *testing.B) {
	f, err := os.Open("testdata/large.tar.gz")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		b.Error(err)
		return
	}
	defer gz.Close()

	var buf bytes.Buffer
	io.Copy(&buf, gz)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compress1X(buf.Bytes())
	}
}

func BenchmarkDecomp(b *testing.B) {
	f, err := os.Open("testdata/large.tar.gz")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		b.Error(err)
		return
	}
	defer gz.Close()

	var buf bytes.Buffer
	io.Copy(&buf, gz)

	cmp := Compress1X(buf.Bytes())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decompress1X(bytes.NewReader(cmp), len(cmp), buf.Len())
	}
}
