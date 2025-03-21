package extendablehash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	BUCKET_SIZE  = 100
	STORAGE_PATH = "./buckets/"
)

type Bucket struct {
	Id         int                    `json:"id"`
	Items      map[string]interface{} `json:"items"`
	LocalDepth int                    `json:"local_depth"`
}

type ExtendableHashTable struct {
	// Директория: отображает индекс (нижние GlobalDepth бит) на указатель на бакет.
	Buckets      map[int]*Bucket
	GlobalDepth  int
	nextBucketId int
}

// NewExtendableHashTable создаёт новую расширяемую хэш‑таблицу и инициализирует 2^GlobalDepth бакетов.
func NewExtendableHashTable() *ExtendableHashTable {
	os.MkdirAll(STORAGE_PATH, os.ModePerm)
	eht := &ExtendableHashTable{
		Buckets:      make(map[int]*Bucket),
		GlobalDepth:  1,
		nextBucketId: 0,
	}
	// Инициализируем 2^GlobalDepth бакетов
	for i := 0; i < (1 << eht.GlobalDepth); i++ {
		b := &Bucket{
			Id:         eht.nextBucketId,
			Items:      make(map[string]interface{}),
			LocalDepth: 1,
		}
		eht.nextBucketId++
		eht.Buckets[i] = b
		eht.saveBucketToFile(b)
	}
	return eht
}

// hash – функция хэширования (алгоритм FNV-1a)
func hash(s string) uint64 {
	var h uint64 = 14695981039346656037
	for _, c := range s {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}

// getBKey вычисляет индекс в директории по ключу с учётом текущей глобальной глубины.
func (eht *ExtendableHashTable) getBKey(key string) int {
	return int(hash(key) & ((1 << eht.GlobalDepth) - 1))
}

// Insert вставляет пару ключ-значение. Если после вставки бакет переполнен,
// запускается обработка переполнения (разделение бакета и/или расширение директории).
func (eht *ExtendableHashTable) Insert(key string, value interface{}) {
	for {
		dirIndex := eht.getBKey(key)
		bucket := eht.loadBucketFromFile(dirIndex)
		bucket.Items[key] = value
		eht.saveBucketToFile(bucket)
		if len(bucket.Items) <= BUCKET_SIZE {
			break
		}
		eht.handleOverflow(dirIndex)
	}
}

// Get возвращает значение по ключу. Если ключ не найден – выводится предупреждение.
func (eht *ExtendableHashTable) Get(key string) (interface{}, bool) {
	dirIndex := eht.getBKey(key)
	bucket := eht.loadBucketFromFile(dirIndex)
	value, exists := bucket.Items[key]
	if !exists {
		fmt.Printf("[WARNING] Key %s not found in bucket %d\n", key, dirIndex)
	}
	return value, exists
}

// expandDirectory расширяет директорию, удваивая число указателей, копируя старые бакеты.
func (eht *ExtendableHashTable) expandDirectory() {
	oldBuckets := make(map[int]*Bucket)
	for k, v := range eht.Buckets {
		oldBuckets[k] = v
	}
	oldGlobalDepth := eht.GlobalDepth
	eht.GlobalDepth++
	eht.Buckets = make(map[int]*Bucket)
	for i := 0; i < (1 << eht.GlobalDepth); i++ {
		eht.Buckets[i] = oldBuckets[i&((1<<oldGlobalDepth)-1)]
	}
}

// handleOverflow проверяет переполнение бакета по индексу dirIndex.
// Если локальная глубина бакета равна глобальной, сначала расширяется директория, затем происходит разделение бакета.
func (eht *ExtendableHashTable) handleOverflow(dirIndex int) {
	bucket := eht.loadBucketFromFile(dirIndex)
	if bucket.LocalDepth == eht.GlobalDepth {
		eht.expandDirectory()
	}
	eht.splitBucket(dirIndex)
}

// splitBucket разделяет бакет, который находится по индексу dirIndex в директории.
// Для разделения вычисляется исходный шаблон (pattern) бакета по его старой локальной глубине (oldLocalDepth).
// Затем локальная глубина старого бакета увеличивается, создаётся новый бакет,
// и обновляются все индексы в директории, для которых (i & ((1 << oldLocalDepth) - 1)) совпадает с pattern:
// если бит на позиции oldLocalDepth равен 1, то этот индекс указывает на новый бакет, иначе – на старый.
func (eht *ExtendableHashTable) splitBucket(dirIndex int) {
	oldBucket := eht.loadBucketFromFile(dirIndex)
	oldLocalDepth := oldBucket.LocalDepth
	// Вычисляем шаблон бакета по его старой локальной глубине.
	pattern := dirIndex & ((1 << oldLocalDepth) - 1)
	// Увеличиваем локальную глубину старого бакета.
	oldBucket.LocalDepth++
	newLocalDepth := oldBucket.LocalDepth
	// Создаём новый бакет с уникальным идентификатором.
	newBucket := &Bucket{
		Id:         eht.nextBucketId,
		Items:      make(map[string]interface{}),
		LocalDepth: newLocalDepth,
	}
	eht.nextBucketId++
	// Обновляем указатели в директории для всех индексов, которые раньше указывали на старый бакет.
	// Для каждого i, если (i & ((1 << oldLocalDepth) - 1)) == pattern, то:
	// - если бит на позиции oldLocalDepth (новый бит) равен 1, то назначаем новый бакет;
	// - иначе оставляем старый бакет.
	for i := 0; i < (1 << eht.GlobalDepth); i++ {
		if (i & ((1 << oldLocalDepth) - 1)) == pattern {
			if (i & (1 << oldLocalDepth)) != 0 {
				eht.Buckets[i] = newBucket
			} else {
				eht.Buckets[i] = oldBucket
			}
		}
	}
	// Перераспределяем ключи: для каждого ключа из старого бакета, если бит на позиции oldLocalDepth в хэше равен 1,
	// переносим запись в новый бакет.
	for key, value := range oldBucket.Items {
		if ((hash(key) >> oldLocalDepth) & 1) == 1 {
			newBucket.Items[key] = value
			delete(oldBucket.Items, key)
		}
	}
	eht.saveBucketToFile(oldBucket)
	eht.saveBucketToFile(newBucket)
}

// saveBucketToFile сохраняет бакет b в файл с именем, основанным на его уникальном Id.
func (eht *ExtendableHashTable) saveBucketToFile(b *Bucket) {
	filePath := fmt.Sprintf("%s%d.json", STORAGE_PATH, b.Id)
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		fmt.Println("Ошибка при маршалинге бакета:", err)
		return
	}
	err = ioutil.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		fmt.Println("Ошибка при сохранении бакета:", err)
	}
}

// loadBucketFromFile загружает бакет, на который ссылается директория по индексу dirIndex,
// используя для имени файла уникальный Id бакета.
func (eht *ExtendableHashTable) loadBucketFromFile(dirIndex int) *Bucket {
	bucket := eht.Buckets[dirIndex]
	filePath := fmt.Sprintf("%s%d.json", STORAGE_PATH, bucket.Id)
	data, err := ioutil.ReadFile(filePath)
	if err == nil {
		var b Bucket
		if err := json.Unmarshal(data, &b); err == nil {
			eht.Buckets[dirIndex] = &b
			return &b
		}
	}
	// Если не удалось прочитать файл, сохраняем текущий бакет и возвращаем его.
	eht.saveBucketToFile(bucket)
	return bucket
}
