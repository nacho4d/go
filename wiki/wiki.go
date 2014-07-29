package main

import (
	"flag"
	"log"
	"net"
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	"fmt"
)

type Page struct {
	Title string
	Body  []byte
}

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

func (page *Page) save() error {
	filename := page.Title + ".txt"
	return ioutil.WriteFile(filename, page.Body, 0600)
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

func renderTemplate(writer http.ResponseWriter, tmpl string, page *Page) {
	tmpl = "tmpl/" + tmpl + ".html"
	fmt.Println("Page title::" + page.Title)
	err := templates.ExecuteTemplate(writer, tmpl, page)
	if err != nil {
		fmt.Println("Couldn't render template: " + tmpl + ". " + err.Error())
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(writer http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		fmt.Println("Page not found, redirecting")
		http.Redirect(writer, request, "/edit/" + title, http.StatusFound)
		return
	}
	renderTemplate(writer, "view", page)
}

func editHandler(writer http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		fmt.Println("Page not found, creating new page:" + title)
		page = &Page{Title: title}
	}
	renderTemplate(writer, "edit", page)
}

func saveHandler(writer http.ResponseWriter, request *http.Request, title string) {
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(writer, request, "/view/" + title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		match := validPath.FindStringSubmatch(request.URL.Path)
		if match == nil {
			http.NotFound(writer, request)
			return
		}
		fn(writer, request, match[2])
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}

	http.ListenAndServe(":8080", nil)
}
