all: gori

gori: gori.go
	go build .

clean:
	rm -f gori

fmt:
	go fmt *.go

install: gori
	cp -f gori /usr/local/bin/gori

test: gori
	go test .

install_deps:
	go get -u github.com/russross/blackfriday
	go get -u github.com/stvp/go-toml-config
	go get -u github.com/tpjg/goriakpbc

deploy: gori
	scp gori maru.thraxil.org:/var/www/gori/
	rsync -ravP media maru.thraxil.org:/var/www/gori/
	ssh maru.thraxil.org "sudo restart gori"

docker: gori
	docker build -t gori:latest .
