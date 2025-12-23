package fs

import (
	"os"
	"testing"

	"github.com/sirrchat/SirrMesh/framework/module"
	"github.com/sirrchat/SirrMesh/internal/storage/blob"
	"github.com/sirrchat/SirrMesh/internal/testutils"
)

func TestFS(t *testing.T) {
	blob.TestStore(t, func() module.BlobStore {
		dir := testutils.Dir(t)
		return &FSStore{instName: "test", root: dir}
	}, func(store module.BlobStore) {
		os.RemoveAll(store.(*FSStore).root)
	})
}
