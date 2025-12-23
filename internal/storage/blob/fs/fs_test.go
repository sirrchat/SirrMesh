package fs

import (
	"os"
	"testing"

	"github.com/mail-chat-chain/sirrmeshd/framework/module"
	"github.com/mail-chat-chain/sirrmeshd/internal/storage/blob"
	"github.com/mail-chat-chain/sirrmeshd/internal/testutils"
)

func TestFS(t *testing.T) {
	blob.TestStore(t, func() module.BlobStore {
		dir := testutils.Dir(t)
		return &FSStore{instName: "test", root: dir}
	}, func(store module.BlobStore) {
		os.RemoveAll(store.(*FSStore).root)
	})
}
