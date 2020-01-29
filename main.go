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
	p, err := loadPage(title)

	// If the requested Page doesn't exist, redirect he client to the edit
	// Page so the content may be created.
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
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

// saveHandler takes the page title (provided in the URL) and the form's only
// field, `Body`, are stored in a new Page. The `save()` method is then called
// to write the data to a file, and the client is redirected to the `/view/`
// page.
func saveHandler(w http.ResponseWriter, r *http.Request) {

	// chop off "/edit/" from the URL and use what is left as the page title
	title := r.URL.Path[len("/edit/"):]

	// fetch provided content in the "body" HTML input field
	body := r.FormValue("body")

	// For the Body field, we must convert `body` to a slice of bytes in order
	// to match the struct field type
	p := &Page{Title: title, Body: []byte(body)}
	p.save()

	// redirect client to the newly saved page
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
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
	http.HandleFunc("/save/", saveHandler)

	// listen on port 8080 on any interface, block until app is terminated
	log.Fatal(http.ListenAndServe(":8000", nil))
}
