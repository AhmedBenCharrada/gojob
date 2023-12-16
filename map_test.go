package gojob_test

import (
	"fmt"
	"gojob"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	m := gojob.Map[string, string]{}
	assert.Equal(t, 0, m.Len())

	m.Remove("k2")

	v, ok := m.Get("k")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", v)

	m.Put("k", "v")
	v, ok = m.Get("k")
	assert.Equal(t, true, ok)
	assert.Equal(t, "v", v)

	v, ok = m.Get("k2")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", v)

	m.Remove("k")
	m.Remove("k2")
	v, ok = m.Get("k")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", v)
}

func TestMap_Range(t *testing.T) {
	m := gojob.Map[string, string]{}
	assert.NotPanics(t, func() {
		m.Range(func(k string, v string) error {
			return nil
		})
	})

	assert.NotPanics(t, func() {
		m := gojob.Map[string, string]{}
		m.Put("k", "v")
		m.Remove("k")
		m.Range(func(k string, v string) error {
			return nil
		})
	})

	m.Put("k", "v")
	m.Put("k2", "v2")
	assert.NotPanics(t, func() {
		type res struct {
			k   string
			v   string
			err error
		}
		ch := make(chan res)
		go func() {
			m.Range(func(k string, v string) error {
				if k == "k2" {
					ch <- res{err: fmt.Errorf("k2")}
					return fmt.Errorf("k2")
				}

				ch <- res{k, v, nil}
				return nil
			})
		}()

		r := <-ch
		assert.Equal(t, "k", r.k)
		assert.Equal(t, "v", r.v)
		r = <-ch
		assert.Error(t, r.err)
	})

}
