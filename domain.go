package main

import (
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

type Page struct {
	Slug     string
	Title    string
	Body     string
	Created  time.Time
	Modified time.Time
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

type PageReadRepository interface {
	FindBySlug(string) (*Page, error)
}

type PageWriteRepository interface {
	Store(*Page) error
}
