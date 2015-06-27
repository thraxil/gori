package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stvp/go-toml-config"
)

type Context struct {
	PageReadRepo  PageReadRepository
	PageWriteRepo PageWriteRepository
	EventStore    EventStore
}

func main() {
	var configFile string
	var loadjson string
	default_conf_file := "./dev.conf"
	if os.Getenv("GORI_CONFIG_FILE") != "" {
		default_conf_file = os.Getenv("GORI_CONFIG_FILE")
	}
	flag.StringVar(&configFile, "config", default_conf_file, "TOML config file")
	flag.StringVar(&loadjson, "loadjson", "", "Load JSON data")
	flag.Parse()

	var (
		port      = config.String("port", "8888")
		media_dir = config.String("media_dir", "media")
	)
	var DB_URL string
	config.Parse(configFile)
	if os.Getenv("GORI_PORT") != "" {
		*port = os.Getenv("GORI_PORT")
	}
	if os.Getenv("GORI_MEDIA_DIR") != "" {
		*media_dir = os.Getenv("GORI_MEDIA_DIR")
	}
	if os.Getenv("GORI_DB_URL") != "" {
		DB_URL = os.Getenv("GORI_DB_URL")
	}

	//	readRepo := NewPGRepo(DB_URL)
	writeRepo := NewPGRepo(DB_URL)
	eventStore := NewPGEventStore(DB_URL)
	readRepo := NewEventStoreReadRepo(eventStore)

	if loadjson != "" {
		log.Println("loading JSON data from", loadjson)
		loadJSON(readRepo, writeRepo, loadjson)
		os.Exit(0)
	}

	var ctx = Context{PageReadRepo: readRepo, PageWriteRepo: writeRepo, EventStore: eventStore}
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.Handle("/", http.RedirectHandler("/page/index/", 302))
	http.HandleFunc("/page/", makeHandler(pageHandler, ctx))
	http.HandleFunc("/edit/", makeHandler(editHandler, ctx))
	http.Handle("/media/", http.StripPrefix("/media/",
		http.FileServer(http.Dir(*media_dir))))
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, Context),
	ctx Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, ctx)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
}

type JsonEntry struct {
	Title    string
	Body     string
	Created  string
	Modified string
}

type JsonFile map[string]JsonEntry

func loadJSON(readRepo PageReadRepository, writeRepo PageWriteRepository, filename string) {
	data, _ := ioutil.ReadFile(filename)
	var entries JsonFile
	err := json.Unmarshal(data, &entries)
	if err != nil {
		log.Println(err)
		return
	}
	for title, entry := range entries {
		slug := slugify(title)
		p, _ := readRepo.FindBySlug(slug)
		p.Title = entry.Title
		p.Body = entry.Body
		modified, err := time.Parse("2006-01-02T15:04:05", entry.Modified)
		if err != nil {
			modified = time.Now()
		}
		p.Modified = modified
		created, err := time.Parse("2006-01-02T15:04:05", entry.Created)
		if err != nil {
			created = modified
		}
		p.Created = created
		writeRepo.Store(p)
	}
}
