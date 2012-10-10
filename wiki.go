package main

import (
	"html/template"
	"net/http"
	"regexp"
	"io/ioutil"
)

const lenPath = len("/view/")

var  (
	// If the templates can't be loaded exit the program (panic).
	templates = template.Must(template.ParseFiles("edit.html", "view.html"))
	// Prevent arbitrary paths being read/written on the server.
	titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")
)

// Page represents a wiki page in memory.
type Page struct {
	Title string
	Body  []byte
}

// Save Page Body to a text file using the Title as the filename.
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// Load the file into memory and return a pointer to the Page.
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handler to view a wiki Page.
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// Handler to edit a wiki Page.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// Handler to save a wiki Page.
// The Page Title (provided in the URL) and the form's only field, Body, 
// are stored in a new Page. The save() method is then called to write the
// data to a file, and the client is redirected to the /view/ page.
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	// The value returned by FormValue is of type string.
	// Convert the value to []byte so it will fit in the Page struct.
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// makeHandler is a validation and error checking wrapper for the handler functions that
// returns a http.HandlerFunc closure.
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the Page title from the Request and call the provided
		// handler 'fn'
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}
