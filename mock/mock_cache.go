package mock

type SetSource struct{}

func (s *SetSource) Build() []interface{} {
	return []interface{}{1, 2, 3, 4, 5}
}

type MapSource struct{}

func (s *MapSource) Build() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"1": 1,
		"2": int64(2),
		"3": "3",
		"4": true,
		"5": 5.0,
	}
}

type ListSource struct{}

func (s *ListSource) Build() []interface{} {
	return []interface{}{0, 1, 2, 3, 4, 5}
}
