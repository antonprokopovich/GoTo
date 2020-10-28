package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const SaveQueueLength = 1000

type URLStore struct {
	urls  map[string]string
	mu    sync.RWMutex
	save chan record
}

type record struct {
	Key string
	URL string
}

// Открывает файл с картой записей ключ-ссылка.
// По окончанию работы с файлом всегда закрывает его.
// В цикле построчно читает файл с помощью энкодера,
// сохраняя записи в струкруту record до тех пор пока
// файл не закончиться или не возникнет ошибка.
func (s *URLStore) load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening URLStore:", err)
		return err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for err == nil {
		var r record
		if err := d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}
	if err == io.EOF {
		return nil
	}
	fmt.Println("Error decoding URLStore:", err)
	return err
}

// NewURLStore инициализирует и возвращает объект хранилища URLStore,
// загружая в него из файла на диске сохраненные ранее записи.
// В отдельной горутине запускает процесс сохранения новых записей.
func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
		save: make(chan record, SaveQueueLength),
	}
	if err := s.load(filename); err != nil {
		log.Println("Error loading ")
	}
	go s.saveLoop(filename)
	return s
}

// saveLoop открывает файл для записи новых ссылок,
// в бесконечном цикле читает новые записи из канала save,
// декодирует их и сохраняет в файл на диск.
func (s *URLStore) saveLoop(filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("URLStore error:", err)
	}
	defer f.Close()

	e := json.NewEncoder(f)
	for {
			r := <- s.save
			if err := e.Encode(&r); err != nil {
				log.Println("URLStore:", err)
		}
	}
}

// Get пытается прочитать длинную ссылку из хранилища по ключу (короткой ссылке).
// При отсутствие записи под переданным ключем вернет пустую строку.
// Блокирует мьютекс на чтение, делая операцию чтения потокобезопасной.
func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

// Set пытается записать ссылку в хранилище под заданным ключем (короткой ссылкой).
// При успешном сохранении возвращает true. Если под заданным ключем уже существует запись,
// то новую запись не сохранияет и возвращает false.
// Блокирует мьютекс на запись, делая операцию записи потокобезопасной.
func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[key]; present {
		return false
	}
	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

// Put генерирует новые ключи на основе текущего размера хранилища до
// тех пор пока не сгенерирует уникальный ключ для сохранения новой ссылки.
// Публикует запись с новым ключем и ссылкой в канал save и возвращает значение ключа.
func (s *URLStore) Put(url string) string {
	for {
		key := genKey(s.Count()) // generate the short URL
		if ok := s.Set(key, url); ok {
			s.save <- record{key, url}
			return key
		}
	}
	panic("shouldn't get here")
}