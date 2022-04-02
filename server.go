package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

// App contains configuration parameters for the web server
type App struct {
	Config      Configuration
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestDump, _ := httputil.DumpRequest(r, true)
	defer r.Body.Close()

	a.DebugLogger.Println("Serving incoming request:\n", string(requestDump))

	response, err := readResponseFromFile(a.Config.ResponseFile)
	if err != nil {
		a.ErrorLogger.Println("Could not read "+a.Config.ResponseFile+" file due to error:", err)
	}
	w.Header().Set("Content-Type", a.Config.ResponseContentType)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func readResponseFromFile(filename string) ([]byte, error) {
	response, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return response, nil
}
