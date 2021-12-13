package simplelru

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
)

func TestLRUWithAccounting(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		if k != string(v.([]byte)) {
			t.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	onAccount := func(k interface{}, v interface{}) int {
		return len(k.(string)) + len(v.([]byte))
	}
	l, err := NewLRUWithAccounting(10, onAccount, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 10; i++ {
		l.Add(fmt.Sprint(i), []byte(fmt.Sprint(i)))
	}
	if l.AccountingSize() != 10 {
		t.Fatalf("bad size: %v", l.AccountingSize())
	}

	if evictCounter != 5 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.Keys() {
		if v, ok := l.Get(k); !ok || string(v.([]byte)) != k || string(v.([]byte)) != fmt.Sprint(i+5) {
			t.Fatalf("bad key: %v", k)
		}
	}
	for i := 0; i < 5; i++ {
		_, ok := l.Get(fmt.Sprint(i))
		if ok {
			t.Fatalf("should be evicted")
		}
	}
	for i := 5; i < 10; i++ {
		_, ok := l.Get(fmt.Sprint(i))
		if !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 5; i < 7; i++ {
		k := fmt.Sprint(i)
		ok := l.Remove(k)
		if !ok {
			t.Fatalf("should be contained")
		}
		ok = l.Remove(k)
		if ok {
			t.Fatalf("should not be contained")
		}
		_, ok = l.Get(k)
		if ok {
			t.Fatalf("should be deleted")
		}
	}

	l.Get(fmt.Sprint(7)) // expect 5 to be last key in l.Keys()

	for i, k := range l.Keys() {
		if (i < 2 && k != fmt.Sprint(i+8)) || (i == 2 && k != "7") {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	if l.Len() != 0 {
		t.Fatalf("bad len: %v", l.Len())
	}
	if l.AccountingSize() != 0 {
		t.Fatalf("bad size: %v", l.Len())
	}
	if _, ok := l.Get(8); ok {
		t.Fatalf("should contain nothing")
	}
}

func TestLRUWithAccounting_update(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k interface{}, v interface{}) {
		evictCounter++
	}
	onAccount := func(k interface{}, v interface{}) int {
		return len(k.(string)) + len(v.([]byte))
	}
	l, err := NewLRUWithAccounting(20, onAccount, onEvicted)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 10; i++ {
		l.Add(fmt.Sprint(i), []byte(fmt.Sprint(i)))
	}

	assert.Equal(t, evictCounter, 0)

	// update
	for i := 0; i < 10; i++ {
		l.Add(fmt.Sprint(i), []byte(fmt.Sprint(i+100)))
	}

	assert.Equal(t, evictCounter, 14)
}
