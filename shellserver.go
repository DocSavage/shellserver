package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var showHelp bool
var showHelp2 bool
var port string
var presentDir string
var shellserverDir string

const shellHtml = "shellserver.html"

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

	http.HandleFunc("/termlib/", frameworkHandler)
	http.HandleFunc("/reveal.js/", frameworkHandler)
	http.HandleFunc("/impress.js/", frameworkHandler)
	http.HandleFunc("/google-io/", googleioHandler)
	http.HandleFunc("/shell", shellHandler)
	http.HandleFunc("/", mainHandler)
	err := src.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Handler for all non-presentation files
func frameworkHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("framework request: %s\n", r.URL)
	filename := filepath.Join(shellserverDir, r.URL.Path)
	http.ServeFile(w, r, filename)
}

// Handler for Google I/O template presentation files
func googleioHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("googleio request: %s\n", r.URL)
	var filename string
	if strings.HasPrefix(r.URL.Path, "/google-io/slide_config.js") ||
		strings.HasPrefix(r.URL.Path, "/google-io/theme/") {

		filename = filepath.Join(presentDir, r.URL.Path)
	} else {
		filename = filepath.Join(shellserverDir, r.URL.Path)
	}
	http.ServeFile(w, r, filename)
}

// Handler for presentation files
func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("presentation request: %s\n", r.URL)
	filename := filepath.Join(presentDir, r.URL.Path)
	http.ServeFile(w, r, filename)
}

// Get a command from POST, parse any query strings that establish
// path or environment variables, execute command and return result.
// Also supports GLOB names like "*.png".
func proxyCommand(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	command := r.FormValue("command")
	args := strings.Split(command, " ")
	if len(args) == 0 {
		fmt.Printf("Bad command (%s)\n", command)
		http.Error(w, "No command given!", http.StatusBadRequest)
		return
	}
	// Expand any arguments with wildcard.
	fullArgs := []string{}
	for _, arg := range args {
		if strings.Contains(arg, "*") {
			matches, err := filepath.Glob(arg)
			if err != nil {
				fmt.Printf("Can't parse glob: %s [%s]\n", arg, err.Error())
			} else {
				fullArgs = append(fullArgs, matches...)
			}
		} else {
			fullArgs = append(fullArgs, arg)
		}
	}

	// Do the command
	fmt.Printf("Passing in arguments: %s\n", fullArgs)
	out, err := exec.Command(fullArgs[0], fullArgs[1:]...).Output()
	if err != nil {
		fmt.Println("Error: ", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		fmt.Println("Output: ", string(out))
		fmt.Fprintln(w, string(out))
	}
}

// Handler for API commands through HTTP
func shellHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		filename := filepath.Join(shellserverDir, shellHtml)
		http.ServeFile(w, r, filename)
	} else if r.Method == "POST" {
		proxyCommand(w, r)
	} else {
		fmt.Fprintln(w, "Got bad command request for shell, not GET or POST!")
	}
}
