package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/russross/blackfriday"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

const (
	HTML_TEMPLATE_EDIT_FILE string = "edit_file.html"
	HTML_TEMPLATE_EDIT_PAGE string = "edit_page.html"
	HTML_TEMPLATE_VIEW_PAGE string = "view_page.html"
)

var (
	TEMPLATES           *template.Template
	HTML_TEMPLATES_EDIT = map[string]string{
		"file": "edit_file.html",
		"page": "edit_page.html",
	}
)

// func getEditTemplate

func init() {
	TEMPLATES = template.Must(template.ParseFiles(HTML_TEMPLATE_EDIT_FILE, HTML_TEMPLATE_EDIT_PAGE, HTML_TEMPLATE_VIEW_PAGE))
}

type Page struct {
	Title string
	Body  template.HTML
	Raw   string
}

type WikiEngine struct{}

func (self *WikiEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if matched, _ := regexp.MatchString(`^/edit/file/*`, r.URL.Path); matched {
		self.editAssetHandler("file", w, r)
		return
	}

	if matched, _ := regexp.MatchString(`^/edit/page/*`, r.URL.Path); matched {
		self.editAssetHandler("page", w, r)
		return
	}

	if matched, _ := regexp.MatchString(`^/file/*`, r.URL.Path); matched {
		self.ViewFileHandler(w, r)
		return
	}

	if matched, _ := regexp.MatchString(`^/page/*`, r.URL.Path); matched {
		self.ViewPageHandler(w, r)
		return
	}

	self.apiResponse(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func (self *WikiEngine) apiResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	if 200 != statusCode {
		logger.Error(message)
		fmt.Fprintf(w, fmt.Sprintf(`{"status":"error","error":"%v"}`, message))
		return
	}
	fmt.Fprintf(w, string(`{"status":"ok"}`))
}

func (self *WikiEngine) loadAsset(atype, path string) (*Page, error) {
	body, err := DB.GetAsset(atype, path)
	return &Page{
		Title: path,
		Body:  template.HTML(blackfriday.MarkdownCommon([]byte(body))),
		Raw:   string(body),
	}, err
}

func (self *WikiEngine) saveAsset(atype, path string, r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return DB.SaveAsset(atype, path, string(body))
}

func (self *WikiEngine) editAssetHandler(atype string, w http.ResponseWriter, r *http.Request) {
	path := strings.Replace(r.URL.Path[1:], "edit/", "", -1)

	if "POST" == r.Method {
		err := self.saveAsset(atype, path, r)
		if err != nil {
			self.apiResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		self.apiResponse(w, http.StatusOK, "ok")
		return
	} else if "DELETE" == r.Method {
		err := DB.DeleteAsset(atype, path)
		if err != nil {
			self.apiResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		self.apiResponse(w, http.StatusOK, "ok")
		return
	}

	p, err := self.loadAsset(atype, path)
	if err != nil {
		logger.Error(err)
		self.renderTemplate(w, HTML_TEMPLATES_EDIT[atype], &Page{Title: path})
		return
	}

	self.renderTemplate(w, HTML_TEMPLATES_EDIT[atype], p)
}

func (self *WikiEngine) ViewFileHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Path[1:]
	p, err := self.loadAsset("file", page)
	if nil != err {
		self.apiResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Fprintf(w, string(p.Raw))
}

func (self *WikiEngine) ViewPageHandler(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Path[1:]
	p, err := self.loadAsset("page", page)
	if nil != err {
		self.apiResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	self.renderTemplate(w, HTML_TEMPLATE_VIEW_PAGE, p)
}

func (self *WikiEngine) renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := TEMPLATES.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
