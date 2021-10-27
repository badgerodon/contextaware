package contextaware

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()

		pr, pw := Pipe()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			n, err := pw.WriteContext(ctx, []byte{1, 2, 3, 4})
			assert.Equal(t, 4, n)
			assert.NoError(t, err)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()

			p := make([]byte, 4)
			n, err := pr.ReadContext(ctx, p)
			assert.Equal(t, []byte{1, 2, 3, 4}, p)
			assert.Equal(t, 4, n)
			assert.NoError(t, err)
		}()
		wg.Wait()
	})

}
