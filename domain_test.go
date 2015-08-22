package main

import (
	"fmt"
	"testing"
)

func TestSetTitle(t *testing.T) {
	p := Page{}
	r := p.SetTitle("")
	if r {
		t.Error("shouldn't be able to set to an empty string")
	}
	r = p.SetTitle("one")
	if !r {
		t.Error("this one should go through")
	}
	r = p.SetTitle("one")
	if r {
		t.Error("shouldn't do anything if i set it to existing title")
	}
}

func TestSetBody(t *testing.T) {
	p := Page{}
	r := p.SetBody("")
	if r {
		t.Error("can't set body to nothing")
	}
	r = p.SetBody("something")
	if !r {
		t.Error("should be able to set the body")
	}
	if p.Body != "something" {
		t.Error("didn't set it")
	}
	r = p.SetBody("something")
	if r {
		t.Error("setting it to existing value should fail")
	}
}

func TestLinkText(t *testing.T) {
	p := Page{}
	p.SetBody("no links")
	if p.LinkText() != p.Body {
		t.Error("nothing should've changed yet")
	}
	p.SetBody("[[a link]]")
	if p.LinkText() != "[a link](/page/a-link/)" {
		t.Error(fmt.Sprintf("didn't handle simple link %s", p.LinkText()))
	}

	p.SetBody("[[   a link   ]]")
	if p.LinkText() != "[a link](/page/a-link/)" {
		t.Error(fmt.Sprintf("didn't handle simple link %s", p.LinkText()))
	}

	p.SetBody("[[---a link---]]")
	if p.LinkText() != "[a link](/page/a-link/)" {
		t.Error(fmt.Sprintf("didn't handle simple link %s", p.LinkText()))
	}

	p.SetBody("[[a link|with title]]")
	if p.LinkText() != "[with title](/page/a-link/)" {
		t.Error(fmt.Sprintf("didn't handle simple link %s", p.LinkText()))
	}
}
