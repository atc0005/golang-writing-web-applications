package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"text/template"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// subexpression 1 is one of edit, save or view
// subexpression 2 is the page title that we are editing, saving or viewing
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

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

// frontPageHandler redirects requests for / to /view/FrontPage
func frontPageHandler(w http.ResponseWriter, r *http.Request) {
	//log.Println("frontPageHandler triggered")
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	return
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {

	p, err := loadPage(title)

	// If the requested Page doesn't exist, redirect the client to the edit
	// Page so the content may be created.
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	// If the Page DOES exist, substitute any Page references to HTML links
	createHTMLPageLinks(p)
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {

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
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {

	// fetch provided content in the "body" HTML input field
	body := r.FormValue("body")

	// For the Body field, we must convert `body` to a slice of bytes in order
	// to match the struct field type
	p := &Page{Title: title, Body: []byte(body)}

	//createWikiPageLinks(p)

	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// redirect client to the newly saved page
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// createWikiPageLinks performs an in-place modification to convert instances
// of HTML page links to [PageName] wiki page links of the same name and
// destination.
func createWikiPageLinks(p *Page) {

	log.Println("createWikiPageLinks called")

	// <a href='/view/ThisPageDoesNotExist'>ThisPageDoesNotExist</a>
	htmlPageLink := regexp.MustCompile("<a href='/view/([a-zA-Z0-9]+)'>[a-zA-Z0-9]+</a>")

	p.Body = htmlPageLink.ReplaceAllFunc(p.Body, func(s []byte) []byte {

		group := htmlPageLink.ReplaceAllString(string(s), "$1")
		pageLink := "[" + group + "]"

		log.Println("createWikiPageLinks - finished generating page link")

		return []byte(pageLink)
	})

}

// createHTMLPageLinks performs an in-place modification to convert instances
// of [PageName] to HTML links of the same name and destination.
func createHTMLPageLinks(p *Page) {

	wikiPageLink := regexp.MustCompile("\\[([a-zA-Z]+)\\]")

	p.Body = wikiPageLink.ReplaceAllFunc(p.Body, func(s []byte) []byte {

		// this appears to replace "[PageName]" with "PageName"
		group := wikiPageLink.ReplaceAllString(string(s), "$1")
		pageLink := "<a href='/view/" + group + "'>" + group + "</a>"

		return []byte(pageLink)
	})

}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {

	// bypass path validation provided by makeHandler (since we know what
	// specific path we're working with) and just call the handler directly
	http.HandleFunc("/", frontPageHandler)

	// loads the page, displays edit form for new page if not existing page
	http.HandleFunc("/view/", makeHandler(viewHandler))

	// displays edit form for existing page, otherwise edit form for new page
	http.HandleFunc("/edit/", makeHandler(editHandler))

	// save the data entered into the edit form
	// TODO
	http.HandleFunc("/save/", makeHandler(saveHandler))

	// listen on port 8080 on any interface, block until app is terminated
	log.Fatal(http.ListenAndServe(":8000", nil))
}
