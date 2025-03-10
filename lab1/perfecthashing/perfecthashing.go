package perfecthashing

import (
	"errors"
	"hash/fnv"
)

type keyValue struct {
	key   string
	value any
}

type PerfectHash struct {
	table []keyValue
	size  int
}

func hashKey(key string, size int) (int, error) {
	if size <= 0 {
		return 0, errors.New("размер должен превышать 0")
	}

	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % size, nil
}

func NewPerfectHash(keys []string, values []any) (*PerfectHash, error) {
	if len(keys) != len(values) {
		return nil, errors.New("keys and values must have the same length")
	}
	if len(keys) == 0 {
		return nil, errors.New("keys cannot be empty")
	}

	size := len(keys) * len(keys)
	table := make([]keyValue, size)
	used := make([]bool, size)

	for i, key := range keys {
		index, err := hashKey(key, size)
		if err != nil {
			return nil, err
		}

		attempts := 0
		for used[index] {
			attempts++
			if attempts > size {
				return nil, errors.New("ошибка при построении таблицы")
			}
			index = (index + attempts*attempts) % size
		}
		used[index] = true
		table[index] = keyValue{key: key, value: values[i]}
	}

	return &PerfectHash{
		table: table,
		size:  size,
	}, nil
}

func (ph *PerfectHash) findKeyValue(key string) (keyValue, bool, error) {
	index, err := hashKey(key, ph.size)
	if err != nil {
		return keyValue{}, false, err
	}
	if ph.table[index].key == key {
		return ph.table[index], true, nil
	}
	return keyValue{}, false, errors.New("key not found")
}

func (ph *PerfectHash) Lookup(key string) (bool, error) {
	_, found, err := ph.findKeyValue(key)
	return found, err
}

func (ph *PerfectHash) GetValueByKey(key string) (any, error) {
	kv, found, err := ph.findKeyValue(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("key not found")
	}
	return kv.value, nil
}

func (ph *PerfectHash) GetAllKeys() []string {
	var keys []string
	for _, kv := range ph.table {
		if kv.key != "" {
			keys = append(keys, kv.key)
		}
	}
	return keys
}

func (ph *PerfectHash) GetAllValues() []any {
	var values []any
	for _, kv := range ph.table {
		if kv.key != "" {
			values = append(values, kv.value)
		}
	}
	return values
}

func (ph *PerfectHash) GetAllKeysValues() []keyValue {
	var kvs []keyValue
	for _, kv := range ph.table {
		if kv.key != "" {
			kvs = append(kvs, kv)
		}
	}
	return kvs
}
