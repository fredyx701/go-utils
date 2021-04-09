package utils

import "sync"

// Set struct set
type Set map[interface{}]struct{}

// NewSet create set with arrs
// init also can use 'make(Set)'
func NewSet(arrs ...[]interface{}) Set {
	var arr []interface{}
	if len(arrs) > 0 {
		arr = arrs[0]
	}
	set := make(Set, len(arr))
	for _, v := range arr {
		set[v] = struct{}{}
	}
	return set
}

// Add 支持 add 多个元素
func (s Set) Add(vs ...interface{}) {
	for _, v := range vs {
		s[v] = struct{}{}
	}
}

// Delete 支持 delete 多个元素
func (s Set) Delete(vs ...interface{}) {
	for _, v := range vs {
		delete(s, v)
	}
}

// AddNX add element if not exists
// if exists, return false
func (s Set) AddNX(v interface{}) bool {
	_, has := s[v]
	if has {
		return false
	}
	s[v] = struct{}{}
	return true
}

func (s Set) Has(v interface{}) bool {
	_, has := s[v]
	return has
}

// List return as []interface{}
func (s Set) List() []interface{} {
	list := make([]interface{}, 0, len(s))
	for k := range s {
		list = append(list, k)
	}
	return list
}

// Subset 判断子集   s contains arr
func (s Set) Subset(arr []interface{}) bool {
	for _, v := range arr {
		if _, has := s[v]; !has {
			return false
		}
	}
	return true
}

// Intersect 取交集
func (s Set) Intersect(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr))
	for _, v := range arr {
		if _, has := s[v]; has {
			result = append(result, v)
		}
	}
	return result
}

// Union 并集
func (s Set) Union(arr []interface{}) []interface{} {
	result := make([]interface{}, 0, len(arr)+len(s))
	for k := range s {
		result = append(result, k)
	}
	for _, v := range arr {
		if _, has := s[v]; !has {
			result = append(result, v)
		}
	}
	return result
}

// Diff 差集
// default return source - target
// isMinuend optional parameters. if true return target - source
func (s Set) Diff(target []interface{}, isMinuend ...bool) (diff []interface{}) {
	arrSet := NewSet(target)
	if len(isMinuend) > 0 && isMinuend[0] {
		diff = make([]interface{}, 0, len(arrSet)/2)
		for k := range arrSet {
			if _, has := s[k]; !has {
				diff = append(diff, k)
			}
		}
	} else {
		diff = make([]interface{}, 0, len(s)/2)
		for k := range s {
			if _, has := arrSet[k]; !has {
				diff = append(diff, k)
			}
		}
	}
	return diff
}

// DiffBoth .
// subtrahend  source - target
// minuend  target - source
func (s Set) DiffBoth(target []interface{}) (subtrahend []interface{}, minuend []interface{}) {
	arrSet := NewSet(target)
	subtrahend = make([]interface{}, 0, len(s)/2)
	for k := range s {
		if _, has := arrSet[k]; !has {
			subtrahend = append(subtrahend, k)
		}
	}
	minuend = make([]interface{}, 0, len(arrSet)/2)
	for k := range arrSet {
		if _, has := s[k]; !has {
			minuend = append(minuend, k)
		}
	}
	return subtrahend, minuend
}

// ....
// ....
// ....
// ------ Safe Set
// ....
// ....
// ....

// SafeSet 使用 RWMutex 实现的 并发安全的 Set
type SafeSet struct {
	sync.RWMutex
	m Set
}

// NewSet .
func NewSafeSet(arrs ...[]interface{}) *SafeSet {
	s := &SafeSet{
		m: NewSet(arrs...),
	}
	return s
}

// Add 支持 add 多个元素
func (s *SafeSet) Add(vs ...interface{}) {
	s.Lock()
	defer s.Unlock()
	s.m.Add(vs...)
}

// Delete 支持 delete 多个元素
func (s *SafeSet) Delete(vs ...interface{}) {
	s.Lock()
	defer s.Unlock()
	s.m.Delete(vs...)
}

// AddNX add element if not exists
// if exists, return false
func (s *SafeSet) AddNX(v interface{}) bool {
	s.Lock()
	defer s.Unlock() // 确保 has add 处于一次事务中
	return s.m.AddNX(v)
}

func (s *SafeSet) Has(v interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	return s.m.Has(v)
}

func (s *SafeSet) Len() int {
	return len(s.m)
}

func (s *SafeSet) Clear() {
	s.Lock()
	defer s.Unlock()
	s.m = make(Set)
}

// List return as []interface{}
func (s *SafeSet) List() []interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.m.List()
}

// Subset 判断子集   s contains arr
func (s *SafeSet) Subset(arr []interface{}) bool {
	s.RLock()
	defer s.RUnlock()
	return s.m.Subset(arr)
}

// Intersect 取交集
func (s *SafeSet) Intersect(arr []interface{}) []interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.m.Intersect(arr)
}

// Union 并集
func (s *SafeSet) Union(arr []interface{}) []interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.m.Union(arr)
}

// Diff 差集
// default return source - target
// isMinuend optional parameters. if true return target - source
func (s *SafeSet) Diff(target []interface{}, isMinuend ...bool) (diff []interface{}) {
	s.RLock()
	defer s.RUnlock()
	return s.m.Diff(target, isMinuend...)
}

// DiffBoth .
// subtrahend  source - target
// minuend  target - source
func (s *SafeSet) DiffBoth(target []interface{}) (subtrahend []interface{}, minuend []interface{}) {
	s.RLock()
	defer s.RUnlock()
	return s.m.DiffBoth(target)
}
