package main

import (
	"encoding/json"
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/russross/blackfriday"
)

type Page struct {
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

func (p *Page) SetTitle(title string) bool {
	if title == "" {
		return false
	}
	if title == p.Title {
		return false
	}
	p.Title = title
	return true
}

func (p *Page) SetBody(body string) bool {
	if body == "" {
		return false
	}
	if body == p.Body {
		return false
	}
	p.Body = body
	return true
}

func (p Page) RenderedBody() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon([]byte(p.LinkText()))))
}

func (p Page) RenderModified() string {
	return p.Modified.Format(time.RFC3339)
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

func (p Page) JSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}

type PageReadRepository interface {
	FindBySlug(string) (*Page, error)
}

type PageWriteRepository interface {
	SetTitle(*Page, string) error
	SetBody(*Page, string) error
}
