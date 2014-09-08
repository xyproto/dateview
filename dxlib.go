package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"os"
	"strings"
)

type ClientState struct {
	username string
	loggedIn bool
}

const (
	version_string = "dxweb 0.2"
	default_server = "default-server:9999"
)

var (
	commandList = []string{"ls", "cd", "login", "l", "test", "exit", "quit", "q"}
)

func ServerWrite(conn net.Conn, line string) (int, error) {
	writer := bufio.NewWriter(conn)
	toWrite := strings.TrimSpace(line)
	if toWrite == "" {
		return 0, errors.New("zero bytes written")
	}
	bytesWritten, err := writer.WriteString(toWrite + "\r\n")
	writer.Flush()
	if err != nil {
		conn.Close()
		return 0, err
	}
	return bytesWritten, err
}

func ServerRead(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	// Received the data from the server
	received, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return "", err
	}
	return received, err
}

func (cs *ClientState) cmdHelper(line string) string {
	// Full line replacments
	switch line {
	case "cmdlist":
		return "list-commands"
	case "l", "login":
		return "login " + cs.username
	case "ls":
		return "list-exams"
	case "reload":
		return "reread-files"
	}
	// Partial replacements
	switch {
	case strings.HasPrefix(line, "cd "):
		return strings.Replace(line, "cd ", "set ", -1)
	}
	// Return the original line if no aliases were recognized
	return line
}

// Modifies the contents of data, for some commands
func postProcessOutput(command string, data string) string {
	switch command {
	case "list-exams":
		data = strings.Replace(data, "not running", " "+"stopped", -1)
		data = strings.Replace(data, ") running", ") "+"running", -1)
	case "list-commands":
		data = strings.Replace(data, ": NIL", "", -1)
		data += "\naliases: cmdlist, l, ls, reload"
	}
	return data
}

func Connect(specified_server string) (cs *ClientState, conn *tls.Conn, server string, err error) {
	cs = &ClientState{username: "", loggedIn: false}

	if specified_server == "" {
		server = default_server

		// Get the server from the environment variable SERVER, if available
		if env_server := os.Getenv("SERVER"); env_server != "" {
			server = env_server
		}
	}

	// Get the username from the commandline or from
	if len(os.Args) >= 2 {
		cs.username = os.Args[1]
	} else {
		// Find the username from the USER environment variable
		if username := os.Getenv("USER"); username != "" {
			cs.username = username
		} else {
			log.Fatalln("No username provided and no USER environment variable found!")
		}
	}

	config := &tls.Config{InsecureSkipVerify: true}
	conn, err = tls.Dial("tcp", server, config)

	// ClientState, connection and connection error (if any)
	return cs, conn, server, err
}

func GetPass() (string, error) {
	return "password", nil
}

func OneCommand(server, line string) string {
	cs, conn, server, err := Connect(server)
	if err != nil {
		log.Fatalln("Unable to connect to " + server + " (" + err.Error() + ")")
	}
	// Close connection at return
	defer conn.Close()

	// Handle exit commands
	if (line == "exit") || (line == "quit") || (line == "q") {
		return "done"
	}

	// Aliases and other replacements
	command := cs.cmdHelper(line)

	// Send the command
	bytesWritten, err := ServerWrite(conn, command)
	if err != nil {
		if bytesWritten == 0 {
			return "done"
		} else {
			log.Println("client: Error when writing to server: " + err.Error())
			return "fail"
		}
	}

	// If sent bytes were 0
	if bytesWritten == 0 {
	}

	// Receive data until "Command done."
	data := ""
	tryToLogIn := false
	for {
		newData, err := ServerRead(conn)
		if err == nil {
			data += newData
		} else {
			log.Fatalln("Error when reading from server: " + err.Error())
		}
		if strings.HasSuffix(strings.TrimSpace(data), "Command done.") {
			if tryToLogIn {
				cs.loggedIn = true
			}
			// Remove the end marker from the output
			data = data[:strings.LastIndex(data, "Command done.")]
			// End of transmission
			break
		} else if strings.HasSuffix(strings.TrimSpace(data), "Password:") {
			tryToLogIn = true

			// Server is expecting a password in return
			password, err := GetPass()
			if err != nil {
				log.Fatalln("client: Error while reading password")
			}

			// Send the password
			ServerWrite(conn, password)
			// Remove the server request from the output
			data = data[:strings.LastIndex(data, "Password:")]
		}
		// Continue receiving
	}
	data = strings.TrimSpace(data)
	if data == "" {
		log.Println("server: Nothing.")
	}
	// Return the output from the server
	return postProcessOutput(command, data)
}

func ServerUp(server string) bool {
	return OneCommand(server, "test") == "Test"
}

func DefaultServerUp() bool {
	return OneCommand("", "test") == "Test"
}
