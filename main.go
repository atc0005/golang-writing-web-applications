package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"gopkg.in/russross/blackfriday.v2"
)

// These directories reside in the same location as the running application
const dataDir string = "data"
const tmplDir string = "tmpl"

// Desired permissions on newly created data or templates directories
const dirPerms os.FileMode = 0700

var templates = template.Must(template.ParseFiles(
	filepath.Join(tmplDir, "edit.html"),
	filepath.Join(tmplDir, "view.html"),
))

// subexpression 1 is one of edit, save or view
// subexpression 2 is the page title that we are editing, saving or viewing
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// Page describes how page data will be stored in memory.
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := filepath.Join(dataDir, p.Title+".txt")
	if !pathExists(dataDir) {
		if err := os.Mkdir(dataDir, dirPerms); err != nil {
			return fmt.Errorf("unable to save page to %q: %s", filename, err)
		}
	}
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := filepath.Join(dataDir, title+".txt")
	//log.Println("filename:", filename)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("error loading page %q: %s", filename, err)
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

	// https://stackoverflow.com/questions/37462126/regex-match-markdown-link
	markdownPageLink := regexp.MustCompile(`(?:__|[*#])|\[(.*?)\]\(.*?\)`)
	wikiPageLink := regexp.MustCompile("\\[([a-zA-Z]+)\\]")

	// ReplaceAllFunc returns a copy of src in which all matches of the Regexp
	// have been replaced by the return value of function repl applied to the
	// matched byte slice. The replacement returned by repl is substituted
	// directly, without using Expand.

	p.Body = wikiPageLink.ReplaceAllFunc(p.Body, func(s []byte) []byte {

		// Don't perform substitution on Markdown page links
		if markdownPageLink.Match(s) {
			log.Println("Skipping match:", string(s))
			return s
		}

		log.Println("Performing substitution against", string(s))

		// this appears to replace "[PageName]" with "PageName"
		group := wikiPageLink.ReplaceAllString(string(s), "$1")
		pageLink := "<a href='/view/" + group + "'>" + group + "</a>"

		return []byte(pageLink)
	})

}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	// TODO:
	//
	// 1) Setup an io.Writer compatible buffer to receive the templates
	// 2) Point templates.ExecuteTemplate at it
	// 3) unsafe := blackfriday.Run(input)
	// 4) html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	// 5) Write "html" to "w"

	//var templateBuffer bytes.Buffer

	if strings.ToLower(tmpl) == "view" {
		p.Body = blackfriday.Run(p.Body)
	}
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// unsafe := blackfriday.Run(templateBuffer.Bytes())
	// html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	// Send converted content to client
	//fmt.Fprint(w, string(output))

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

// pathExists confirms that the specified path exists
func pathExists(path string) bool {

	// Make sure path isn't empty
	if strings.TrimSpace(path) == "" {
		log.Println("path is empty string")
		return false
	}

	// https://gist.github.com/mattes/d13e273314c3b3ade33f
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		//log.Println("path found")
		return true
	}

	return false

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
