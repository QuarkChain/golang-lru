golang-lru
==========

This provides the `lru` package which implements a
- fixed-size
- customized-accounting
LRU cache. It is based on the cache in Groupcache.

Documentation
=============

Full docs are available on [Godoc](http://godoc.org/github.com/hashicorp/golang-lru)

Example
=======

Using the customized-space-accounting LRU is very simple:

```go
onAccount := func(k interface{}, v interface{}) int {
    return len(k.(string)) + len(v.([]byte))
}

l, _ := NewLRUWithAccounting(10, onAccount, nil)


for i := 0; i < 10; i++ {
    l.Add(fmt.Sprint(i), []byte(fmt.Sprint(i)))
}
```
