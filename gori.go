package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/russross/blackfriday"
	"github.com/stvp/go-toml-config"
)

type Page struct {
	Slug     string
	Title    string
	Body     string
	Created  time.Time
	Modified time.Time
}

func getPage(db *sql.DB, slug string) (*Page, error) {
	stmt, err := db.Prepare(
		"select title, body, created, modified from pages where slug = $1")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	row := stmt.QueryRow(slug)
	if row == nil {
		return nil, nil
	}
	var title string
	var body string
	var created time.Time
	var modified time.Time

	row.Scan(&title, &body, &created, &modified)

	p := Page{slug, title, body, created, modified}
	return &p, nil
}

func (p *Page) Create(db *sql.DB) error {
	stmt, err := db.Prepare(
		"insert into pages (slug, created) values ($1, $2)")

	if err != nil {
		log.Println(err)
		return err
	}

	now := time.Now()
	_, err = stmt.Exec(p.Slug, now)
	return err
}

func (p *Page) SaveAs(db *sql.DB, slug string) error {
	stmt, err := db.Prepare(
		"update pages set title = $1, body = $2, modified = $3 where slug = $4")

	if err != nil {
		log.Println(err)
		return err
	}

	now := time.Now()
	_, err = stmt.Exec(p.Title, p.Body, now, slug)
	return err
}

func (p Page) RenderedBody() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(p.LinkText()))))
}

func makeLink(s string) string {
	// s should look like '[[Page Title]]'
	// or [[Page Title|link text]]
	// we turn those into
	// [Page Title](/page/page-title/)
	// or
	// [link text](/page/page-title/)
	// respectively
	s = strings.Trim(s, "[]- ") // get rid of the delimiters
	title := s
	link := "/page/" + slugify(s) + "/"
	if strings.Index(s, "|") != -1 {
		parts := strings.SplitN(s, "|", 2)
		page_title := strings.Trim(parts[0], " ")
		link_text := strings.Trim(parts[1], " ")
		title = link_text
		link = "/page/" + slugify(page_title) + "/"
	}
	return "[" + title + "](" + link + ")"
}

func (p Page) LinkText() string {
	pattern, _ := regexp.Compile(`(\[\[\s*[^\|\]]+\s*\|?\s*[^\]]*\s*\]\])`)
	return pattern.ReplaceAllStringFunc(p.Body, makeLink)
}

func slugify(s string) string {
	s = strings.Trim(s, " \t\n\r-")
	s = strings.Replace(s, " ", "-", -1)
	s = strings.ToLower(s)
	return s
}

type Context struct {
	DB *sql.DB
}

func main() {
	var configFile string
	default_conf_file := "./dev.conf"
	if os.Getenv("GORI_CONFIG_FILE") != "" {
		default_conf_file = os.Getenv("GORI_CONFIG_FILE")
	}
	flag.StringVar(&configFile, "config", default_conf_file, "TOML config file")
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

	db, err := sql.Open("postgres", DB_URL)

	if err != nil {
		log.Println("can't open database")
		log.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	var ctx = Context{db}
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

type PageResponse struct {
	Title    string
	Slug     string
	Body     template.HTML
	Modified time.Time
}

func pageHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 400)
		return
	}
	slug := parts[2]
	if slug == "" {
		http.Error(w, "bad request", 400)
		return
	}
	page, err := getPage(ctx.DB, slug)
	if err != nil {
		log.Println(err)
		http.Error(w, "error retrieving page", 500)
		return
	}
	if page.Title == "" {
		http.Redirect(w, r, "/edit/"+slug+"/", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	pr := PageResponse{
		Title:    page.Title,
		Slug:     slugify(page.Title),
		Body:     page.RenderedBody(),
		Modified: page.Modified,
	}
	t, _ := template.New("page").Parse(page_view_template)
	t.Execute(w, pr)
}

const page_view_template = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>{{.Title}}</title>
 <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="gori">
  <meta name="author" content="anders pearson">
    <link href="/media/bootstrap/css/bootstrap.css" rel="stylesheet">
    <link href="/media/bootstrap/css/bootstrap-responsive.css" rel="stylesheet">
    <link href="/media/css/main.css" rel="stylesheet">
    <link type="text/css" rel="stylesheet" href="/media/main.css" />
 <script src="/media/js/jquery-1.7.2.min.js"></script>
<script src="http://html5shim.googlecode.com/svn/trunk/html5.js"></script>
</head>
<body>
<div class="navbar navbar-fixed-top navbar-inverse">
    <div class="navbar-inner">
      <div class="container">
        <ul class="nav">
          <li><a class="brand" href="/"><i class="icon-home icon-white"></i></a></li>
        </ul>
      </div>
    </div>
</div>
<div class="container" id="outer-container">
<p class="muted pull-right">Last Modified: <b>{{.Modified}}</b></p>
<h1>{{.Title}} <small><a href="/edit/{{.Slug}}/"><i class="icon-edit"></i></a></small></h1>
{{.Body}}
</div>
<script type="text/javascript" src="http://platform.twitter.com/widgets.js"></script>
<script src="/media/bootstrap/js/bootstrap.js"></script>
</body>
</html>
`

type EditPageResponse struct {
	Title    string
	Slug     string
	Existing bool
	Body     template.HTML
}

func deslug(s string) string {
	s = strings.Replace(s, "-", " ", -1)
	s = strings.Title(s)
	return s
}

func editHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) < 3 {
		http.Error(w, "bad request", 400)
		return
	}
	slug := parts[2]
	if slug == "" {
		http.Error(w, "bad request", 400)
		return
	}
	page, err := getPage(ctx.DB, slug)
	if err != nil {
		log.Println(err)
		http.Error(w, "error retrieving page", 500)
		return
	}

	if r.Method == "POST" {
		if r.FormValue("create") == "true" {
			page.Create(ctx.DB)
		}
		page.Body = r.FormValue("body")
		page.Title = r.FormValue("title")
		page.Modified = time.Now()
		page.SaveAs(ctx.DB, slug)
		http.Redirect(w, r, "/page/"+slug+"/", http.StatusFound)
	} else {
		// just show the edit form
		w.Header().Set("Content-Type", "text/html")
		title := page.Title
		var existing = false
		if page.Title == "" {
			title = deslug(slug)
			existing = false
		}
		t, _ := template.New("edit").Parse(page_edit_template)
		t.Execute(w, EditPageResponse{
			Title:    title,
			Slug:     slugify(page.Title),
			Existing: existing,
			Body:     template.HTML(page.Body),
		})
	}
}

const page_edit_template = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8" />
<title>Edit {{.Title}}</title>
 <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="gori">
  <meta name="author" content="anders pearson">
    <link href="/media/bootstrap/css/bootstrap.css" rel="stylesheet">
    <link href="/media/bootstrap/css/bootstrap-responsive.css" rel="stylesheet">
    <link href="/media/css/main.css" rel="stylesheet">
    <link type="text/css" rel="stylesheet" href="/media/main.css" />
 <script src="/media/js/jquery-1.7.2.min.js"></script>
<script src="http://html5shim.googlecode.com/svn/trunk/html5.js"></script>
</head>
<body>
<div class="navbar navbar-fixed-top navbar-inverse">
    <div class="navbar-inner">
      <div class="container">
        <ul class="nav">
          <li><a class="brand" href="/"><i class="icon-home icon-white"></i></a></li>
        </ul>
      </div>
    </div>
</div>
<div class="container" id="outer-container">

<form action="." method="post">
<fieldset>
<legend>Edit {{.Title}}</legend>
<input type="text" name="title" value="{{.Title}}" placeholder="title" class="input-block-level"/>
<textarea name="body" rows="30" class="input-block-level">{{.Body}}</textarea>
{{ if .Existing }}
<a class="btn" href="/edit/{{.Slug}}/">cancel</a>
{{ else }}
<input type="hidden" name="create" value="true" />
<a class="btn" href="/page/index/">cancel</a>
{{ end }}
<input class="btn btn-primary" type="submit" value="save">
</form>
</div>
<script type="text/javascript" src="http://platform.twitter.com/widgets.js"></script>
<script src="/media/bootstrap/js/bootstrap.js"></script>
</body>
</html>
`
