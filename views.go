package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

type PageResponse struct {
	Title    string
	Slug     string
	Body     template.HTML
	Modified string
}

func pageHandler(w http.ResponseWriter, r *http.Request, ctx Context) {
	log.Println("pageHandler", r.URL.String())
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
	page, err := ctx.PageReadRepo.FindBySlug(slug)
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
		Modified: page.RenderModified(),
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
	page, err := ctx.PageReadRepo.FindBySlug(slug)
	if err != nil {
		log.Println(err)
		http.Error(w, "error retrieving page", 500)
		return
	}

	if r.Method == "POST" {
		page.Slug = slug
		ctx.PageWriteRepo.SetTitle(page, r.FormValue("title"))
		ctx.PageWriteRepo.SetBody(page, r.FormValue("body"))
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
