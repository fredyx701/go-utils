package utils

import (
	"sync"
	"time"
)

// MapSource map data source
// Build() failed return nil.  判断 nil, 不会更新数据.  SetSource  ListSource 同
type MapSource interface {
	Build() map[interface{}]interface{}
}

// Map map 缓存
type Map struct {
	sync.RWMutex
	cache     map[interface{}]interface{}
	expiredAt int64
	expire    int64
	source    MapSource
}

// NewMap 创建 map 缓存
// expire 缓存保留时间
// opts[0]  check duration   默认 1h
func NewMap(source MapSource, expire time.Duration, opts ...interface{}) *Map {
	obj := &Map{
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
	go obj.check(duration)
	return obj
}

// check cache map
func (m *Map) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if m.expiredAt < next.Unix() {
			m.Lock()
			if m.expiredAt < next.Unix() {
				m.cache = make(map[interface{}]interface{})
			}
			m.Unlock()
		}
	}
}

// build cache
func (m *Map) build() {
	maps := m.source.Build()
	if maps != nil {
		m.cache = make(map[interface{}]interface{}, len(maps))
		for k, v := range maps {
			m.cache[k] = v
		}
	}
	m.expiredAt = time.Now().Unix() + m.expire
}

func (m *Map) checkBuild() {
	now := time.Now().Unix()
	if m.expiredAt < now {
		m.Lock()
		if m.expiredAt < now { // 二次确认   for  parallel build
			m.build()
		}
		m.Unlock()
	}
}

// Get get value
func (m *Map) Get(key interface{}) (interface{}, bool) {
	m.checkBuild()
	m.RLock()
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
	cache     map[interface{}]struct{}
	expiredAt int64
	expire    int64
	source    SetSource
}

// NewSet 创建 set 缓存
func NewSet(source SetSource, expire time.Duration, opts ...interface{}) *Set {
	obj := &Set{
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
	go obj.check(duration)
	return obj
}

// check cache set
func (s *Set) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if s.expiredAt < next.Unix() {
			s.Lock()
			if s.expiredAt < next.Unix() {
				s.cache = make(map[interface{}]struct{})
			}
			s.Unlock()
		}
	}
}

// build cache
func (s *Set) build() {
	slice := s.source.Build()
	if slice != nil {
		s.cache = make(map[interface{}]struct{}, len(slice))
		for _, v := range slice {
			s.cache[v] = struct{}{}
		}
	}
	s.expiredAt = time.Now().Unix() + s.expire
}

func (s *Set) checkBuild() {
	now := time.Now().Unix()
	if s.expiredAt < now {
		s.Lock()
		if s.expiredAt < now {
			s.build()
		}
		s.Unlock()
	}
}

// Has .
func (s *Set) Has(key interface{}) bool {
	s.checkBuild()
	s.RLock()
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
	s.checkBuild()
	s.RLock()
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
	s.checkBuild()
	s.RLock()
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
	s.checkBuild()
	s.RLock()
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

// ListSource  list data source
type ListSource interface {
	Build() []interface{}
}

// List list 缓存
type List struct {
	sync.RWMutex
	cache     []interface{}
	expiredAt int64
	expire    int64
	source    ListSource
}

// NewList 创建 list 缓存
func NewList(source ListSource, expire time.Duration, opts ...interface{}) *List {
	obj := &List{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make([]interface{}, 0),
	}
	duration := time.Hour // 默认 1h
	if len(opts) > 0 {
		param, ok := opts[0].(time.Duration)
		if !ok {
			panic("params must be time.Duration")
		}
		duration = param
	}
	go obj.check(duration)
	return obj
}

// check cache list
func (s *List) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		if s.expiredAt < next.Unix() {
			s.Lock()
			if s.expiredAt < next.Unix() {
				s.cache = make([]interface{}, 0)
			}
			s.Unlock()
		}
	}
}

// build cache
func (s *List) build() {
	slice := s.source.Build()
	if slice != nil {
		s.cache = make([]interface{}, len(slice))
		s.cache = slice
	}
	s.expiredAt = time.Now().Unix() + s.expire
}

func (s *List) checkBuild() {
	now := time.Now().Unix()
	if s.expiredAt < now {
		s.Lock()
		if s.expiredAt < now {
			s.build()
		}
		s.Unlock()
	}
}

// Get 获取原 slice
func (s *List) Get() []interface{} {
	s.checkBuild()
	return s.cache
}

// Copy 获取副本
func (s *List) Copy() []interface{} {
	s.checkBuild()
	s.RLock()
	slice := make([]interface{}, len(s.cache))
	copy(slice, s.cache)
	s.RUnlock()
	return slice
}

// Length .
func (s *List) Length() int {
	return len(s.cache)
}

// CacheSource map data source
// Build() failed return nil, 不会更新数据.  SetSource  ListSource 同
type CacheSource interface {
	Build(key interface{}, opts ...interface{}) interface{}
}

// Store kv 缓存
type Store struct {
	sync.RWMutex
	cache  map[interface{}]*storeEelment
	expire int64
	source CacheSource
}

type storeEelment struct {
	sync.RWMutex
	expiredAt int64
	value     interface{}
}

// NewStore 创建 kv 缓存
func NewStore(source CacheSource, expire time.Duration, opts ...interface{}) *Store {
	obj := &Store{
		expire: int64(expire.Seconds()),
		source: source,
		cache:  make(map[interface{}]*storeEelment),
	}
	duration := time.Hour // 默认 1h
	if len(opts) > 0 {
		param, ok := opts[0].(time.Duration)
		if !ok {
			panic("params must be time.Duration")
		}
		duration = param
	}
	go obj.check(duration)
	return obj
}

// check cache map
func (m *Store) check(duration time.Duration) {
	c := time.Tick(duration)
	for next := range c {
		now := next.Unix()
		for k, v := range m.cache {
			if v.expiredAt < now {
				m.Lock()
				if v.expiredAt < now {
					delete(m.cache, k)
				}
				m.Unlock()
			}
		}
	}
}

// build cache
func (m *Store) build(val *storeEelment, key interface{}, opts ...interface{}) {
	result := m.source.Build(key, opts)
	if result != nil {
		val.value = result
	}
	val.expiredAt = time.Now().Unix() + m.expire // 延续之前的值 or 保留 nil 值
}

func (m *Store) checkBuild(key interface{}, opts ...interface{}) {
	// check exist
	val, has := m.cache[key]
	if !has {
		m.Lock()
		val, has = m.cache[key]
		if !has { // check value
			val = &storeEelment{
				expiredAt: 0,
				value:     nil, // 预创建，避免 不存在的数据 频繁 build
			}
			m.cache[key] = val
		}
		m.Unlock()
	}
	// check expireAt
	now := time.Now().Unix()
	if val.expiredAt < now {
		val.Lock()
		if val.expiredAt < now { // check value
			m.build(val, key, opts...)
		}
		val.Unlock()
	}
}

// Get get value
func (m *Store) Get(key interface{}, opts ...interface{}) (interface{}, bool) {
	m.checkBuild(key, opts...)
	m.RLock()
	val, has := m.cache[key]
	m.RUnlock()
	if val.value == nil {
		has = false
	}
	return val.value, has
}

// Size .
func (m *Store) Size() int {
	return len(m.cache)
}
