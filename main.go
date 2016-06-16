package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Page struct {
	Title   string
	Content []byte
}

func (p *Page) save() error {
	filename := "blog_data/" + p.Title + ".txt"
	log.Println("save file: ", filename)
	return ioutil.WriteFile(filename, p.Content, 0600)
}

func main() {
	dir, _ := os.Stat("blog_data")
	if dir == nil || !dir.IsDir() {
		log.Println("The director 'blog_data' is not exists. Create it now.")
		os.Mkdir("blog_data", 0777)
	}

	// handle the list page
	http.HandleFunc("/list", listHandle)
	// handle the detail page
	http.HandleFunc("/view/", viewHandle)
	// handle the add or edit page
	http.HandleFunc("/edit/", editHandle)
	// handle the save data
	http.HandleFunc("/save/", saveHandle)

	// handle the static files like css, js
	http.HandleFunc("/public/", staticHandler)

	log.Println("Start server with port 8080")
	http.ListenAndServe(":8080", nil)

}

func listHandle(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("blog_data")
	if err != nil {
		log.Println("empty list.")
		renderTemplate(w, "list", nil)
	} else {
		length1 := len(files)

		ps := make([]*Page, length1)
		for i := 0; i < length1; i++ {
			name := strings.Split(files[i].Name(), ".")[0]
			ps[i] = &Page{Title: name, Content: nil}
		}
		t, _ := template.ParseFiles("template/list.html")
		t.Execute(w, ps)
	}
}

func viewHandle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		log.Println("Article is not exist, create it.")
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/save/"):]
	content := r.FormValue("content")
	p := &Page{Title: title, Content: []byte(content)}
	err := p.save()
	if err != nil {
		fmt.Println(err)
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[1:]
	log.Println("file path: ", filePath)
	content, err := ioutil.ReadFile(filePath)

	if err != nil {
		log.Println("not found static file : ", filePath)
		fmt.Fprintf(w, "404")
		return
	}
	var ctype string
	if strings.Contains(filePath, ".css") {
		ctype = "text/css"
	} else if strings.Contains(filePath, ".js") {
		ctype = "text/javascript"
	}
	// TODO process more static file. such as picture
	// it seems there is a package include this funcion
	w.Header().Set("Content-Type", ctype)
	fmt.Fprintf(w, "%s\n", content)
}

func loadPage(title string) (*Page, error) {
	filename := "blog_data/" + title + ".txt"
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Content: content}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles("template/" + tmpl + ".html")
	t.Execute(w, p)
}
