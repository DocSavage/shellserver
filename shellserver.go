package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	
	"code.google.com/p/go.net/websocket"
)

var showHelp bool
var showHelp2 bool
var port string

var presentDir string
var shellserverDir string
var shellDir string

var currentCommand *command

const shellHtml = "shellserver.html"

const helpMessage = `
shellserver is a web server + shell proxy ideally suited to run
presentations using tools like reveal.js

	usage: shellserver [options]

For more information, visit:
http://github.com/DocSavage/shellserver/README.md
`

// Communicator can talk channels and abstracts clients speaking
// through a websocket and a goroutine running a command.
type Communicator interface {
	reader()
	writer()
	outchannel() chan string
	close()
}

type command struct {
	cmd string
	
	// Buffered channel of outbound messages.
	send chan string
	
	stdin io.WriteCloser
	stdout io.ReadCloser
}

func (c *command) outchannel() chan string {
	if c == nil {
		fmt.Println("Tried to reference nil command.outchannel")
		return nil
	}
	return c.send
}

func (c *command) close() {
	// nothing needed
}
 
func (c *command) reader() {
	bufreader := bufio.NewReader(c.stdout)
	buf := make([]byte, 0, 10000)
	for {
		n, err := bufreader.Read(buf)
		if err != nil {
			fmt.Println("command reader() error:", err.Error())
			break
		}
		message = strings.TrimSpace(string(buffer))
		fmt.Println("Received data from command:", message)
		h.broadcast <- message
	}
	c.close()
}
 
func (c *command) writer() {
	for message := range c.send {
		fmt.Println("command writer() printing:", message)
		_, err := fmt.Fprintf(c.stdin, message)
		if err != nil {
			fmt.Println("command writer() error:", err.Error())
			break
		}
	}
	c.close()
}
 
func doCommand(cmd string) {
	if currentCommand != nil {
		fmt.Println("Closing previous command:", currentCommand.cmd)
		currentCommand.close()
	}
	cmd = strings.TrimSpace(cmd)
	fmt.Println("Launching new command:", cmd)
	currentCommand = &command{send: make(chan string, 256), cmd: cmd}
	pCmd := processCommand(cmd)
	var err error
	currentCommand.stdin, err = pCmd.StdinPipe()
	if err != nil {
		fmt.Println("Can't connect to stdin of command:", cmd)
	}
	currentCommand.stdout, err = pCmd.StdoutPipe()
	if err != nil {
		fmt.Println("Can't connect to stdout of command:", cmd)
	}
	err = pCmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err.Error())
		return
	}
	h.register <- currentCommand
	defer func() { h.unregister <- currentCommand }()
	go currentCommand.writer()
	currentCommand.reader()
	pCmd.Wait()
	currentCommand = nil
}

// Websocket hub system courtesy of Gary Burd
// https://gist.github.com/garyburd/1316852
type connection struct {
	// The websocket connection.
	ws *websocket.Conn
 
	// Buffered channel of outbound messages.
	send chan string
}

func (c *connection) outchannel() chan string {
	return c.send
}

func (c *connection) close() {
	if c != nil && c.ws != nil {
		c.ws.Close()	
	}
}
 
func (c *connection) reader() {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			break
		}
		fmt.Println("websocket read:", message)
		if currentCommand != nil {
			fmt.Println("reader() sending message to current program:",
				currentCommand.cmd)
			h.broadcast <- message
		} else {
			go doCommand(message)
		}
	}
	c.close()
}
 
func (c *connection) writer() {
	for message := range c.send {
		fmt.Println("websocket writing:", message)
		err := websocket.Message.Send(c.ws, message)
		if err != nil {
			fmt.Println("websocket writer() error:", err.Error())
			break
		}
	}
	c.close()
}
 
func wsHandler(ws *websocket.Conn) {
	c := &connection{send: make(chan string, 256), ws: ws}
	h.register <- c
	defer func() { h.unregister <- c }()
	go c.writer()
	c.reader()
}

type hub struct {
	// Registered connections.
	connections map[Communicator]bool
 
	// Inbound messages from the connections.
	broadcast chan string
 
	// Register requests from the connections.
	register chan Communicator
 
	// Unregister requests from connections.
	unregister chan Communicator
}
 
var h = hub{
	broadcast:   make(chan string),
	register:    make(chan Communicator),
	unregister:  make(chan Communicator),
	connections: make(map[Communicator]bool),
}
 
func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.connections[c] = true
		case c := <-h.unregister:
			delete(h.connections, c)
			c.close()
		case m := <-h.broadcast:
			for c := range h.connections {
				select {
				case c.outchannel() <- m:
				default:
					delete(h.connections, c)
					c.close()
				}
			}
		}
	}
}


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
	flag.StringVar(&shellDir, "cd", currentDir,
		"Directory for running shell commands")
}

func main() {
	flag.Parse()

	if showHelp || showHelp2 {
		// Print local DVID help
		fmt.Println(helpMessage)
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	} else {
		err := os.Chdir(shellDir)
		if err != nil {
			fmt.Println("Error trying to change working directory to:", shellDir)
		}
		go h.run()
		serveHttp("localhost:" + port)
	}
}

// Listen and serve HTTP requests using address and don't let stay-alive
// connections hog goroutines for more than an hour.
// See for discussion: 
// http://stackoverflow.com/questions/10971800/golang-http-server-leaving-open-goroutines
func serveHttp(address string) {

	fmt.Printf("Launching shellserver on %s ...\n", address)
	fmt.Printf("Launching websocket listener on %s ...\n", address+"/socket/")

	src := &http.Server{
		Addr:        address,
		ReadTimeout: 1 * time.Hour,
	}

	http.HandleFunc("/termlib/", frameworkHandler)
	http.HandleFunc("/reveal.js/", frameworkHandler)
	http.HandleFunc("/impress.js/", frameworkHandler)
	http.HandleFunc("/google-io/", googleioHandler)
	http.HandleFunc("/shell", shellHandler)
	http.Handle("/socket/", websocket.Handler(wsHandler))
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

func processCommand(command string) (cmd *exec.Cmd) {
	args := strings.Split(command, " ")

	if len(args) == 0 {
		fmt.Printf("Bad command (%s)\n", command)
		return
	}
	// If this is "cd", then change working directory.
	if args[0] == "cd" {
		if len(args) < 2 {
			fmt.Println("'cd' must be followed with new directory!")
		} else {
			err := os.Chdir(args[1])
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Switched directory to", args[1])
			}
			shellDir = args[1]
		}
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

	// Return the command
	cmd = exec.Command(fullArgs[0], fullArgs[1:]...)
	return
}

// Get a command from POST, parse any query strings that establish
// path or environment variables, execute command and return result.
// Also supports GLOB names like "*.png".
func proxyCommand(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	command := r.FormValue("command")
	args := strings.Split(command, " ")

	if len(args) == 0 {
		fmt.Fprintf(w, "Bad command (%s)\n", command)
		return
	}
	// If this is "cd", then change working directory.
	if args[0] == "cd" {
		if len(args) < 2 {
			fmt.Fprintln(w, "'cd' must be followed with new directory!")
		} else {
			err := os.Chdir(args[1])
			if err != nil {
				fmt.Fprintln(w, err.Error())
			} else {
				fmt.Fprintln(w, "Switched directory to", args[1])
			}
			shellDir = args[1]
		}
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

	// Check for "&" at end to signify asynchronous command like server starts.
	lastArg := len(fullArgs) - 1
	runBackground := false
	if fullArgs[lastArg] == "&" {
		runBackground = true
		fullArgs = fullArgs[:lastArg]
	}

	// Do the command
	cmd := exec.Command(fullArgs[0], fullArgs[1:]...)
	var out []byte
	var err error
	if runBackground {
		err = cmd.Start()
		out = []byte(fmt.Sprintf("Ran background job: %s\n", command))
	} else {
		out, err = cmd.Output()
	}
	if err != nil {
		fmt.Println("Error: ", err.Error())
		fmt.Fprintln(w, err.Error())
	} else {
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
