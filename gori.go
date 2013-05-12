package main

import (
	"fmt"
	"github.com/russross/blackfriday"
	"github.com/tpjg/goriakpbc"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Page struct {
	Title    string ""
	Body     string ""
	Created  string ""
	Modified string ""
	riak.Model
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
	Client *riak.Client
}

func main() {
	// 8087 is the port for protocol buffers
	client := riak.New("128.59.152.25:8087")
	err := client.Connect()
	if err != nil {
		fmt.Println("Cannot connect, is Riak running?")
		return
	}

	var ctx = Context{client}
	http.Handle("/", http.RedirectHandler("/page/index/", 302))
	http.HandleFunc("/page/", makeHandler(pageHandler, ctx))
	http.HandleFunc("/edit/", makeHandler(editHandler, ctx))
	http.Handle("/media/", http.StripPrefix("/media/",
		http.FileServer(http.Dir("media"))))
	log.Fatal(http.ListenAndServe(":8888", nil))
	client.Close()
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
	Modified string
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
	var page Page
	err := ctx.Client.Load("riakipage", slug, &page)
	if err != nil {
		// it seems that the riak client likes to return warnings
		// that i don't understand. At some point, I should
		// figure out what it's complaining about and do something
		// with this err instead of just ignoring it.
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
	var page Page
	err := ctx.Client.Load("riakipage", slug, &page)
	if err != nil {
		// it seems that the riak client likes to return warnings
		// that i don't understand. At some point, I should
		// figure out what it's complaining about and do something
		// with this err instead of just ignoring it.
	}

	if r.Method == "POST" {
		if r.FormValue("create") == "true" {
			ctx.Client.New("riakipage", slug, &page)
		}
		page.Body = r.FormValue("body")
		page.Title = r.FormValue("title")
		t := time.Now()
		page.Modified = t.Format(time.RFC3339)
		page.SaveAs(slug)
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
