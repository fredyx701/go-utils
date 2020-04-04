package utils

import (
	"sync"
	"time"
)

// MapSource map data source
type MapSource interface {
	Build() map[interface{}]interface{}
}

// Map map 缓存
type Map struct {
	sync.RWMutex
	cache      map[interface{}]interface{}
	expireTime int64
	expire     int64
	source     MapSource
}

// NewMap 创建 map 缓存
func NewMap(source MapSource, expire time.Duration, opts ...interface{}) *Map {
	cache := &Map{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make(map[interface{}]interface{}),
	}
	duration := time.Hour // 默认 1h
	if len(opts) > 0 {
		param, ok := opts[0].(time.Duration)
		if !ok {
			panic("params must be time.Duration")
		}
		duration = param
	}
	go cache.check(duration)
	return cache
}

// check cache map
func (m *Map) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if m.expireTime < next.Unix() {
			m.Lock()
			if m.expireTime < next.Unix() {
				m.cache = make(map[interface{}]interface{})
			}
			m.Unlock()
		}
	}
}

// Build build cache
func (m *Map) Build() {
	m.Lock()
	maps := m.source.Build()
	length := len(maps)
	if length != 0 {
		m.cache = make(map[interface{}]interface{}, length)
		for k, v := range maps {
			m.cache[k] = v
		}
	}
	m.expireTime = time.Now().Unix() + m.expire
	m.Unlock()
}

// Get get value
func (m *Map) Get(key interface{}) (interface{}, bool) {
	now := time.Now().Unix()
	m.RLock()
	if m.expireTime < now {
		m.RUnlock()
		m.Build()
		m.RLock()
	}
	val, has := m.cache[key]
	m.RUnlock()
	return val, has
}

// GetBool .
func (m *Map) GetBool(key interface{}) bool {
	val, has := m.Get(key)
	if has {
		return val.(bool)
	}
	return false
}

// GetFloat64 .
func (m *Map) GetFloat64(key interface{}) float64 {
	val, has := m.Get(key)
	if has {
		return val.(float64)
	}
	return 0
}

// GetInt64 .
func (m *Map) GetInt64(key interface{}) int64 {
	val, has := m.Get(key)
	if has {
		return val.(int64)
	}
	return 0
}

// GetInt .
func (m *Map) GetInt(key interface{}) int {
	val, has := m.Get(key)
	if has {
		return val.(int)
	}
	return 0
}

// GetString .
func (m *Map) GetString(key interface{}) string {
	val, has := m.Get(key)
	if has {
		return val.(string)
	}
	return ""
}

// Size .
func (m *Map) Size() int {
	return len(m.cache)
}

// Set set value
func (m *Map) Set(key interface{}, val interface{}) {
	m.Lock()
	m.cache[key] = val
	m.Unlock()
}

// Delete delete value
func (m *Map) Delete(key interface{}) {
	m.Lock()
	delete(m.cache, key)
	m.Unlock()
}

// SetSource  set data source
type SetSource interface {
	Build() []interface{}
}

// Set set 缓存
type Set struct {
	sync.RWMutex
	cache      map[interface{}]struct{}
	expireTime int64
	expire     int64
	source     SetSource
}

// NewSet 创建 set 缓存
func NewSet(source SetSource, expire time.Duration, opts ...interface{}) *Set {
	cache := &Set{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make(map[interface{}]struct{}),
	}
	duration := time.Hour // 默认 1h
	if len(opts) > 0 {
		param, ok := opts[0].(time.Duration)
		if !ok {
			panic("params must be time.Duration")
		}
		duration = param
	}
	go cache.check(duration)
	return cache
}

// check cache set
func (s *Set) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if s.expireTime < next.Unix() {
			s.Lock()
			if s.expireTime < next.Unix() {
				s.cache = make(map[interface{}]struct{})
			}
			s.Unlock()
		}
	}
}

// Build build cache
func (s *Set) Build() {
	s.Lock()
	slice := s.source.Build()
	length := len(slice)
	if length != 0 {
		s.cache = make(map[interface{}]struct{}, length)
		for _, v := range slice {
			s.cache[v] = struct{}{}
		}
	}
	s.expireTime = time.Now().Unix() + s.expire
	s.Unlock()
}

// Has .
func (s *Set) Has(key interface{}) bool {
	now := time.Now().Unix()
	s.RLock()
	if s.expireTime < now {
		s.RUnlock()
		s.Build()
		s.RLock()
	}
	_, has := s.cache[key]
	s.RUnlock()
	return has
}

// Size .
func (s *Set) Size() int {
	return len(s.cache)
}

// Add .
func (s *Set) Add(key interface{}) {
	s.Lock()
	s.cache[key] = struct{}{}
	s.Unlock()
}

// Delete .
func (s *Set) Delete(key interface{}) {
	s.Lock()
	delete(s.cache, key)
	s.Unlock()
}

// Intersect 取交集
func (s *Set) Intersect(arr []interface{}) []interface{} {
	now := time.Now().Unix()
	s.RLock()
	if s.expireTime < now {
		s.RUnlock()
		s.Build()
		s.RLock()
	}
	result := make([]interface{}, 0, len(arr))
	for _, v := range arr {
		_, has := s.cache[v]
		if has {
			result = append(result, v)
		}
	}
	s.RUnlock()
	return result
}

// Union 取并集
func (s *Set) Union(arr []interface{}) []interface{} {
	now := time.Now().Unix()
	s.RLock()
	if s.expireTime < now {
		s.RUnlock()
		s.Build()
		s.RLock()
	}
	result := make([]interface{}, 0, len(arr)+len(s.cache))
	for k := range s.cache {
		result = append(result, k)
	}
	for _, v := range arr {
		_, has := s.cache[v]
		if !has {
			result = append(result, v)
		}
	}
	s.RUnlock()
	return result
}

// Diff 取差集
func (s *Set) Diff(arr []interface{}) []interface{} {
	now := time.Now().Unix()
	s.RLock()
	if s.expireTime < now {
		s.RUnlock()
		s.Build()
		s.RLock()
	}
	arrSet := make(map[interface{}]struct{}, len(arr))
	for _, v := range arr {
		arrSet[v] = struct{}{}
	}
	result := make([]interface{}, 0, len(s.cache))
	for k := range s.cache {
		_, has := arrSet[k]
		if !has {
			result = append(result, k)
		}
	}
	s.RUnlock()
	return result
}
