//package main
//
//import (
//	"fmt"
//	"lab1/extendiblehashing"
//)
//
//func main() {
//	// Создаём хеш-таблицу без фиксированного лимита.
//	// Изначально globalDepth=1, localDepth=1, значит "виртуальная емкость" = 2^1 = 2.
//
//	eh := extendiblehashing.NewExtendibleHash()
//
//	// Добавляем элементы
//	keys := []int{9, 19, 29, 39, 49, 60, 71, 88, 93, 104, 115, 230, 351, 473, 569, 777}
//	for _, key := range keys {
//		eh.Insert(key)
//	}
//
//	// Выводим структуру
//	eh.Print()
//
//	// Проверяем поиск
//	fmt.Println("Поиск 35:", eh.Find(93))
//	fmt.Println("Поиск 100:", eh.Find(230))
//
//	// Удаляем элемент
//	eh.Delete(35)
//	fmt.Println("\nПосле удаления 35:")
//	eh.Print()
//}

package main

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

type PerfectHash struct {
	size      int
	buckets   [][]string
	hashFuncs []struct{ a, b, m int }
	table     [][]string
}

func NewPerfectHash(keys []string) *PerfectHash {
	rand.Seed(time.Now().UnixNano())

	ph := &PerfectHash{
		size:      len(keys),
		buckets:   make([][]string, len(keys)),
		hashFuncs: make([]struct{ a, b, m int }, len(keys)),
		table:     make([][]string, len(keys)),
	}

	// Распределение ключей по бакетам
	for _, key := range keys {
		index := ph.hash1(key) % ph.size
		ph.buckets[index] = append(ph.buckets[index], key)
	}

	// Создание хеш-функций для каждого бакета
	for i, bucket := range ph.buckets {
		if len(bucket) == 0 {
			continue
		}

		m := len(bucket) * len(bucket)
		var a, b int
		success := false

		for !success {
			a = rand.Intn(1000) + 1
			b = rand.Intn(1000) + 1
			tempTable := make([]string, m)
			success = true

			for _, key := range bucket {
				idx := ph.hash2(a, b, m, key)
				if tempTable[idx] != "" && tempTable[idx] != key {
					success = false
					break
				}
				tempTable[idx] = key
			}
		}

		ph.hashFuncs[i] = struct{ a, b, m int }{a, b, m}
		ph.table[i] = make([]string, m)

		for _, key := range bucket {
			idx := ph.hash2(a, b, m, key)
			ph.table[i][idx] = key
		}
	}

	return ph
}

func (ph *PerfectHash) hash1(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32())
}

func (ph *PerfectHash) hash2(a, b, m int, key string) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return (a*int(h.Sum64()) + b) % 2147483647 % m
}

func (ph *PerfectHash) Get(key string) (string, bool) {
	bucketIdx := ph.hash1(key) % ph.size
	if len(ph.buckets[bucketIdx]) == 0 {
		return "", false
	}

	hParams := ph.hashFuncs[bucketIdx]
	idx := ph.hash2(hParams.a, hParams.b, hParams.m, key)

	if idx >= len(ph.table[bucketIdx]) {
		return "", false
	}

	value := ph.table[bucketIdx][idx]
	if value == "" {
		return "", false
	}
	return value, true
}

func main() {
	keys := []string{"apple", "banana", "cherry", "date", "fig"}
	ph := NewPerfectHash(keys)

	// Проверка
	for _, key := range keys {
		if value, ok := ph.Get(key); ok {
			fmt.Printf("%-8s -> %s\n", key, value)
		} else {
			fmt.Printf("%-8s -> not found\n", key)
		}
	}

	// Проверка несуществующего ключа
	if _, ok := ph.Get("mango"); !ok {
		fmt.Println("mango -> not found")
	}
}
