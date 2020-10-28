package main

import (
	"flag"
	"fmt"
	"net/http"
)

const addForm = `
<html><body>
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
</html></body>
`

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
	dataFile = flag.String("store", "store.json", "data store file name")
	hostname = flag.String("host", "localhost:8080", "host name and port")
)

var store *URLStore
func main() {
	flag.Parse()
	store = NewURLStore(*dataFile)

	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(*listenAddr, nil)
}

// Redirect обрабатывает GET-запросы коротких ссылок, перенаправляя
// на соответвующий оригинальный адресс.
func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)
	if url == "" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

// Add обрабатывает POST-запросы на добавление новой длинной ссылки в хранилище.
// При успешном добавлении выводит ключ, под которым ссылка была сохранена.
func Add(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprint(w, addForm)
		return
	}
	key := store.Put(url)

	fmt.Fprintf(w, "%s", key)
}
