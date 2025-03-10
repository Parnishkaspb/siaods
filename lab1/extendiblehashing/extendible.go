package extendiblehashing

import (
	"fmt"
	"hash/fnv"
)

type keyValue struct {
	key   string
	value any
}

// Bucket - структура для хранения пар ключ-значение
type Bucket struct {
	table   []keyValue
	ld      int
	maxSize int
}

// ExtendibleHash - структура для расширяемого хеширования.
type ExtendibleHash struct {
	directories []*Bucket
	gd          int
}

// hashKey - первичное хеширование
func hashKey(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32())
}

// secondHash - вторичное хеширование (Необходим для XOR)
func secondHash(key string) int {
	h := fnv.New32a()
	h.Write([]byte("salt-" + key)) // Дополняем строку для уникальности
	return int(h.Sum32())
}

// finalHash - объединяем два хеша через XOR
func finalHash(key string) int {
	return hashKey(key) ^ secondHash(key)
}

// NewExtendibleHash создает новую хеш-таблицу.
func NewExtendibleHash(initialGD int, maxSize int) *ExtendibleHash {
	size := 1 << initialGD // 2^GD директорий
	directories := make([]*Bucket, size)

	for i := range directories {
		directories[i] = &Bucket{
			table:   make([]keyValue, 0, maxSize), // Изначально пустой
			ld:      initialGD,
			maxSize: maxSize,
		}
	}

	return &ExtendibleHash{
		directories: directories,
		gd:          initialGD,
	}
}

// InsertKey добавляет ключ в таблицу.
func (eh *ExtendibleHash) InsertKey(key string, value any) error {
	h := finalHash(key)
	index := h & ((1 << eh.gd) - 1)
	bucket := eh.directories[index]

	// Если ключ уже есть — обновляем
	for i, kv := range bucket.table {
		if kv.key == key {
			bucket.table[i].value = value
			return nil
		}
	}

	// Если есть место — вставляем
	if len(bucket.table) < bucket.maxSize {
		bucket.table = append(bucket.table, keyValue{key, value})
		return nil
	}

	// Бакет переполнен -> split
	eh.splitBucket(index)

	// 🔹 Пересчитываем индекс, но теперь проверяем, есть ли место
	newIndex := h & ((1 << eh.gd) - 1)
	newBucket := eh.directories[newIndex]

	if len(newBucket.table) < newBucket.maxSize {
		newBucket.table = append(newBucket.table, keyValue{key, value})
		return nil
	}

	for i := 0; i < len(eh.directories); i++ {
		if len(eh.directories[i].table) < eh.directories[i].maxSize {
			eh.directories[i].table = append(eh.directories[i].table, keyValue{key, value})
			return nil
		}
	}

	return fmt.Errorf("не удалось вставить ключ %s после разбиения", key)
}

func (eh *ExtendibleHash) expandDirectory() {
	oldSize := 1 << eh.gd
	eh.gd++
	newSize := 1 << eh.gd

	newDirs := make([]*Bucket, newSize)

	// Дублируем ссылки на старые бакеты в новой таблице
	for i := 0; i < newSize; i++ {
		newDirs[i] = eh.directories[i%oldSize]
	}

	eh.directories = newDirs
}

func (eh *ExtendibleHash) splitBucket(index int) {
	bucket := eh.directories[index]

	// Проверка: переполнение бакета
	if len(bucket.table) < bucket.maxSize {
		return
	}

	// Локальная равна глобальной -> расширяем таблицу
	if bucket.ld == eh.gd {
		eh.expandDirectory()
	}

	// Создание нового бакета с увеличенной локальной глубиной
	newBucket := &Bucket{
		table:   make([]keyValue, 0, bucket.maxSize),
		ld:      bucket.ld + 1,
		maxSize: bucket.maxSize,
	}

	// Увеличиваем локальную глубину старого бакета
	bucket.ld++

	oldKeys := bucket.table
	bucket.table = make([]keyValue, 0, bucket.maxSize)

	// Перераспределение ключей между старым и новым бакетом
	for _, kv := range oldKeys {
		newIndex := finalHash(kv.key) & ((1 << bucket.ld) - 1)
		if newIndex == index {
			bucket.table = append(bucket.table, kv)
		} else {
			newBucket.table = append(newBucket.table, kv)
		}
	}

	// Обновление ссылок в Директории
	for i := range eh.directories {
		if (i & ((1 << (bucket.ld - 1)) - 1)) == index {
			if (i & (1 << (bucket.ld - 1))) == 0 {
				eh.directories[i] = bucket
			} else {
				eh.directories[i] = newBucket
			}
		}
	}
}

// Lookup проверяет наличие ключа в таблице.
func (eh *ExtendibleHash) Lookup(key string) (bool, any) {
	h := finalHash(key) // XOR-хеш для поиска
	index := h & ((1 << eh.gd) - 1)
	bucket := eh.directories[index]

	for _, kv := range bucket.table {
		if kv.key == key {
			return true, kv.value
		}
	}

	return false, nil
}

// DeleteKey удаляет ключ из хеш-таблицы.
func (eh *ExtendibleHash) DeleteKey(key string) error {
	h := finalHash(key)
	index := h & ((1 << eh.gd) - 1)
	bucket := eh.directories[index]

	for i, kv := range bucket.table {
		if kv.key == key {
			bucket.table = append(bucket.table[:i], bucket.table[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("ключ %s не найден", key)
}
