package utils

import (
	"testing"
	"time"

	"github.com/FredyXue/go-utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	sets := NewSet(&mock.SetSource{}, time.Second, time.Second*2)

	assert.Equal(t, true, sets.Has(1))
	assert.Equal(t, false, sets.Has(10))
	assert.Equal(t, 5, sets.Size())

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, sets.Size())

	assert.Equal(t, true, sets.Has(1))
	assert.Equal(t, 5, sets.Size())

	arr := []interface{}{4, 5, 6, 7, 8}
	res1 := sets.Intersect(arr)
	res2 := sets.Union(arr)
	res3 := sets.Diff(arr)
	assert.ElementsMatch(t, []interface{}{4, 5}, res1)
	assert.ElementsMatch(t, []interface{}{1, 2, 3, 4, 5, 6, 7, 8}, res2)
	assert.ElementsMatch(t, []interface{}{1, 2, 3}, res3)
}

func TestList(t *testing.T) {
	list := NewList(&mock.ListSource{}, time.Second, time.Second*2)
	slice := list.Copy()
	assert.EqualValues(t, 0, slice[0])
	assert.EqualValues(t, 6, len(slice))

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, list.Length())

	slice2 := list.Get()
	assert.EqualValues(t, 1, slice2[1])
}

func TestMap(t *testing.T) {
	maps := NewMap(&mock.MapSource{}, time.Second, time.Second*2)
	assert.Equal(t, 1, maps.GetInt("1"))
	assert.Equal(t, int64(2), maps.GetInt64("2"))
	assert.Equal(t, "3", maps.GetString("3"))
	assert.Equal(t, true, maps.GetBool("4"))
	assert.Equal(t, 5.0, maps.GetFloat64("5"))

	v1, has := maps.Get(10)
	assert.Equal(t, false, has)
	assert.Equal(t, nil, v1)
	assert.Equal(t, 5, maps.Size())

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, maps.Size())

	assert.Equal(t, 1, maps.GetInt("1"))
	assert.Equal(t, 5, maps.Size())
}

func TestStore(t *testing.T) {
	store := NewStore(&mock.StoreSource{}, time.Second, time.Second*2)
	val, _ := store.Get(1)
	store.Get(2)
	arr := val.([]interface{})
	assert.EqualValues(t, 2, arr[1])
	assert.Equal(t, 2, store.Size())

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, store.Size())

	val, _ = store.Get(1)
	arr = val.([]interface{})
	assert.EqualValues(t, 2, arr[1])
	assert.Equal(t, 1, store.Size())
}
