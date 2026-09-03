package starter

type Reader struct {
	primary map[string]int64
	cache   map[string]int64
}

func New(primary map[string]int64) *Reader {
	return &Reader{primary: primary, cache: make(map[string]int64)}
}

func (r *Reader) Balance(id string) int64 {
	if value, ok := r.cache[id]; ok {
		return value
	}
	value := r.primary[id]
	r.cache[id] = value
	return value
}
