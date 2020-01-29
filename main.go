package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

// Page describes how page data will be stored in memory.
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, _ := loadPage(title)
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {

	// chop off "/edit/" from the URL and use what is left as the page title
	title := r.URL.Path[len("/edit/"):]

	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, "edit", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func main() {

	// loads the page, displays edit form for new page if not existing page
	http.HandleFunc("/view/", viewHandler)

	// displays edit form for existing page, otherwise edit form for new page
	http.HandleFunc("/edit/", editHandler)

	// save the data entered into the edit form
	// TODO
	//http.HandleFunc("/save/", saveHandler)

	// listen on port 8080 on any interface, block until app is terminated
	log.Fatal(http.ListenAndServe(":8000", nil))
}
