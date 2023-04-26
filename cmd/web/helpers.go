package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"text/template"
	"time"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clienrError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clienrError(w, http.StatusNotFound)
}

func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	return td
}

func (app *application) render(w http.ResponseWriter, r *http.Request, files []string, td *templateData) {
	ts, err := template.ParseFiles(files...)
	if err != nil {

		app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}

	err = ts.Execute(w, app.addDefaultData(td, r))
	if err != nil {

		app.serverError(w, err)
		ErrorHandler(w, http.StatusText(http.StatusInternalServerError), 500)
		return
	}
}

type ErrorBody struct {
	Message string
	Code    int
}

func ErrorHandler(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	// fmt.Println(code)
	errorBody := ErrorBody{Message: message, Code: code}
	html, err := template.ParseFiles("./ui/templates/error.html")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err = html.Execute(w, errorBody); err != nil {
		// fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
}
