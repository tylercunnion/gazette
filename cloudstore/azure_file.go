package cloudstore

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/Azure/azure-sdk-for-go/storage"
)

const (
	// Maximum size of an Azure blob storage block.
	maxAzureBlockSize = 104857600
)

type azureFile struct {
	container *storage.Container
	blob      *storage.Blob

	readStream io.ReadCloser

	currOffset int64

	buf *bufio.Writer
}

func (a *azureFile) Close() error {
	var err error
	if a.readStream != nil {
		if err = a.readStream.Close(); err != nil {
			return err
		}
	}

	if a.buf != nil {
		if err = a.buf.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (a *azureFile) Read(p []byte) (n int, err error) {
	defer func() { a.currOffset += n }()

	// If necessary, open the read stream.
	if a.readStream == nil {
		var stream, err = a.blob.Get(nil)
		if err != nil {
			return 0, err
		}
		a.readStream = stream
	}
	return a.readStream.Read(p)
}

func (a *azureFile) Seek(offset int64, whence int) (newOffset int64, err error) {
	defer func() {
		a.currOffset = newOffset
	}()

	// Dispose of any existing stream.
	if a.readStream != nil {
		a.readStream.Close()
		a.readStream = nil
	}

	var absOffset int64
	switch whence {
	case io.SeekStart:
		absOffset = offset
	case io.SeekCurrent:
		absOffset = offset + a.currOffset
	case io.SeekEnd:
		// TODO(tyler): Read properties and get full length
		return 0, fmt.Errorf("seek end not implemented")
	default:
		return 0, fmt.Errorf("whence not recognized")
	}

	var blobRange = &Storage.BlobRange{
		Start: uint64(absOffset),
	}

	var stream, err = a.blob.GetRange(&storage.GetBlobRangeOptions{
		Range: blobRange,
	})
	if err != nil {
		return 0, err
	}
	a.readStream = stream
	return absOffset, nil
}

func (a *azureFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *azureFile) Stat() (os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *azureFile) Write(p []byte) (n int, err error) {
	// Lazy initialize interior writer.
	if a.buf == nil {
		a.buf = bufio.NewWriterSize(&azureWriter{
			blob: a.blob,
		}, maxAzureBlockSize)
	}

	return a.buf.Write(p)
}

// ContentSignature is a representation of the file's data, ideally
// a content sum or ETag (in the case of cloud storage providers).
// Calling this should not require a calculation that reads the whole file.
func (a *azureFile) ContentSignature() (string, error) {
	if err := a.blob.GetProperties(nil); err != nil {
		return "", err
	}
	return a.blob.Properties.Etag, nil
}

type azureWriter struct {
	blob *storage.Blob
	part int
}

func (w *azureWriter) Write(p []byte) (n int, err error) {
	var blockId = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", w.part)))
	var err = w.blob.PutBlock(blockId, p, nil)
	if err != nil {
		return 0, err
	}
	w.part++
	return len(p), nil
}
