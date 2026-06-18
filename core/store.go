package core

var store map[string]string

func Init() {
	store = make(map[string]string)
}

func Get(key string) string {
	val, exists := store[key]
	if !exists {
		return "-1"
	}
	return val
}

func Put(key, value string) {
	store[key] = value
}
