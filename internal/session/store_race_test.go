package session

import (
	"sync"
	"testing"
)

func TestStore_ConcurrentListPaginatedAndSave(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sess, err := store.Create("claude", "/tmp")
			if err != nil {
				return
			}
			sess.AddTag("alpha")
			_ = store.Save(sess)
		}()
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.ListPaginated(&ListFilter{Tag: "alpha"})
		}()
	}

	wg.Wait()
}
