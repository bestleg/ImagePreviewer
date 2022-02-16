package cache

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)
		c.Set("first", 1)
		c.Set("second", "second")
		c.Set("third", "3rd")

		val, ok := c.Get("third") // third first second
		require.Equal(t, "3rd", val)
		require.True(t, ok)

		val, ok = c.Get("second") // second third first
		require.Equal(t, "second", val)
		require.True(t, ok)

		c.Set("4th", "4th") // 4th second third
		val, ok = c.Get("4th")
		require.Equal(t, "4th", val)
		require.True(t, ok)

		_, ok = c.Get("first") // deleted
		require.False(t, ok)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			nBig, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
			require.NoError(t, err, "rand error")
			c.Get(Key(strconv.FormatInt(nBig.Int64(), 10)))
		}
	}()

	wg.Wait()
}
