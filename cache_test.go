package utils

import (
	"testing"
	"time"

	"github.com/FredyXue/go-utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	sets := NewSet(&mock.SetSource{}, time.Second*2, time.Millisecond*500)

	assert.Equal(t, true, sets.Has(1))
	assert.Equal(t, false, sets.Has(10))
	assert.Equal(t, 5, sets.Size())

	// preload
	time.Sleep(time.Second)
	sets.Build()
	assert.Equal(t, 5, sets.Size())

	time.Sleep(time.Second) // preload extend expiredAt
	assert.Equal(t, 5, sets.Size())

	time.Sleep(time.Second) // clear
	assert.Equal(t, 0, sets.Size())

	// build
	assert.Equal(t, true, sets.Has(1))
	assert.Equal(t, 5, sets.Size())

	// intersect
	arr := []interface{}{4, 5, 6, 7, 8}
	res1 := sets.Intersect(arr)
	res2 := sets.Union(arr)
	res3 := sets.Diff(arr)
	assert.ElementsMatch(t, []interface{}{4, 5}, res1)
	assert.ElementsMatch(t, []interface{}{1, 2, 3, 4, 5, 6, 7, 8}, res2)
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, res3)
}

func TestList(t *testing.T) {
	list := NewList(&mock.ListSource{}, time.Second*2, time.Millisecond*500)
	slice := list.Copy()
	assert.EqualValues(t, 0, slice[0])
	assert.EqualValues(t, 6, len(slice))

	// preload
	time.Sleep(time.Second)
	list.Build()
	assert.Equal(t, 6, list.Length())

	time.Sleep(time.Second)
	assert.Equal(t, 6, list.Length())

	time.Sleep(time.Second)
	assert.Equal(t, 0, list.Length())

	slice = list.Get()
	assert.EqualValues(t, 1, slice[1])
}

func TestMap(t *testing.T) {
	maps := NewMap(&mock.MapSource{}, time.Second*2, time.Millisecond*500)
	assert.Equal(t, 1, maps.GetInt("1"))
	assert.Equal(t, int64(2), maps.GetInt64("2"))
	assert.Equal(t, "3", maps.GetString("3"))
	assert.Equal(t, true, maps.GetBool("4"))
	assert.Equal(t, 5.0, maps.GetFloat64("5"))

	v1, has := maps.Get(10)
	assert.Equal(t, false, has)
	assert.Equal(t, nil, v1)
	assert.Equal(t, 5, maps.Size())

	time.Sleep(time.Second)
	maps.Build()
	assert.Equal(t, 5, maps.Size())

	time.Sleep(time.Second)
	assert.Equal(t, 5, maps.Size())

	time.Sleep(time.Second)
	assert.Equal(t, 0, maps.Size())

	assert.Equal(t, 1, maps.GetInt("1"))
	assert.Equal(t, 5, maps.Size())
}

func TestStore(t *testing.T) {
	store := NewStore(&mock.StoreSource{}, time.Second*2, time.Millisecond*500)
	val, _ := store.Get(1)

	time.Sleep(time.Second) // 1s
	store.Get(2)
	arr := val.([]interface{})
	assert.EqualValues(t, 2, arr[1])
	assert.Equal(t, 2, store.Size())

	time.Sleep(time.Second) // 2s
	assert.Equal(t, 1, store.Size())
	time.Sleep(time.Second) // 3s check
	assert.Equal(t, 0, store.Size())

	val, _ = store.Get(1)
	arr = val.([]interface{})
	assert.EqualValues(t, 2, arr[1])
	assert.Equal(t, 1, store.Size())

	time.Sleep(time.Second) // 4s
	store.Build(false, 1)   // prebuild 1
	time.Sleep(time.Second) // 5s  check
	assert.Equal(t, 1, store.Size())
	time.Sleep(time.Second) // 6s  check
	assert.Equal(t, 0, store.Size())
}
