package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var showHelp bool
var showHelp2 bool
var port string
var presentDir string
var shellserverDir string

const helpMessage = `
shellserver is a web server + shell proxy ideally suited to run
presentations using tools like reveal.js

	usage: shellserver [options]

For more information, visit:
http://github.com/DocSavage/shellserver/README.md
`

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get current directory:", err.Error())
	}

	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showHelp2, "help", false, "Show help message")

	flag.StringVar(&port, "port", "6789", "Port number to use for server")
	flag.StringVar(&presentDir, "present", currentDir,
		"Directory that holds presentation HTML")
	flag.StringVar(&shellserverDir, "shellserver", currentDir,
		"The shellserver working directory")
}

func main() {
	flag.Parse()

	if showHelp || showHelp2 {
		// Print local DVID help
		fmt.Println(helpMessage)
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	} else {
		serveHttp("localhost:" + port)
	}
}

// Listen and serve HTTP requests using address and don't let stay-alive
// connections hog goroutines for more than an hour.
// See for discussion: 
// http://stackoverflow.com/questions/10971800/golang-http-server-leaving-open-goroutines
func serveHttp(address string) {

	fmt.Printf("Launching shellserver on %s ...\n", address)

	src := &http.Server{
		Addr:        address,
		ReadTimeout: 1 * time.Hour,
	}

	http.HandleFunc("/reveal.js/", frameworkHandler)
	http.HandleFunc("/google-io/", frameworkHandler)
	http.HandleFunc("/shellserver/", shellHandler)
	http.HandleFunc("/", mainHandler)
	src.ListenAndServe()
}

// Handler for all non-presentation files
func frameworkHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Join(shellserverDir, r.URL.Path)
	http.ServeFile(w, r, filename)
}

// Handler for presentation files
func mainHandler(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Join(presentDir, r.URL.Path)
	http.ServeFile(w, r, filename)
}

// Handler for API commands through HTTP
func shellHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>ShellServer API Handler ...</h1>")

	const lenPath = len("/shellserver/")
	url := r.URL.Path[lenPath:]

	fmt.Fprintln(w, "<p>Processing", url, "</p>")
}
