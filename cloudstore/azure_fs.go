package cloudstore

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type azureFs struct {
	container *storage.Container
}

func newAzureFsWithKey(accountName, accountKey, containerName string) (*azureFs, error) {
	var client, err = storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		return nil, err
	}
	var blobSvc = client.GetBlobService()
	var container = blobSvc.GetContainerReference(containerName)
	return &azureFs{
		container: container,
	}
}

func (fs *azureFs) Open(name string) (http.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// Close is a no-op.
func (fs *azureFs) Close() error {
	return nil
}

func (fs *azureFs) CopyAtomic(to File, from io.Reader) (n int64, err error) {
	return 0, fmt.Errorf("not implemented")
}

func (fs *azureFs) MkdirAll(name string, perm os.FileMode) error {
	return fmt.Errorf("not implemented")
}

func (fs *azureFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	//TODO(tyler): make file modes do stuff
	var blob = fs.container.GetBlobReference(name)
	return &azureFile{
		container: fs.container,
		blob:      blob,
	}
}

func (fs *azureFs) ProducesAuthorizedURL() bool {
	return false
}

func (fs *azureFs) Remove(name string) error {
	var blob = fs.container.GetBlobReference(name)
	var deleted, err = blob.DeleteIfExists(nil)
	if err != nil {
		return err
	} else if !deleted {
		return os.ErrNotExist
	}
	return nil
}

func (fs *azureFs) ToURL(name, method string, validFor time.Duration) (*url.URL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fs *azureFs) Walk(root string, walkFn filepath.WalkFunc) error {
	return fmt.Errorf("not implemented")
}
