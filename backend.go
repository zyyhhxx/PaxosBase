package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

//Data structure for the data
type Homework struct {
	Name string
	Desc string
}

type FrontendMessage struct {
	Operation   string
	ID          int
	NewHomework Homework
}

type BackendMessage struct {
	Success    bool
	Homeworks  []Homework
	ErrMessage string
	Homework   Homework
}

//Global variable for the data
var homeworks = make([]Homework, 0, 5)

/*
The main loop of the program
Input: none
Output: none
*/
func main() {

	//Set flags
	portNumPtr := flag.Int("listen", 8090, "The port number the server will listen at")
	flag.Parse()

	//Parse any command-line arguments
	portNum := *portNumPtr

	//Initialize the data
	initializeData()

	//Run the server
	run(portNum)
}

/*
This function initializes the data used for the server
Input: none
Output: none
*/
func initializeData() {
	homeworks = append(homeworks, Homework{"hw1", "Getting to know Go"})
	homeworks = append(homeworks, Homework{"proj1", "Step 1 to the grand project"})
	homeworks = append(homeworks, Homework{"hw2", "Getting to know Go again"})
	homeworks = append(homeworks, Homework{"proj2", "Step 2 to the grand project"})
}

/*
This function runs the backend server
Input: an int portnumber
Output: none
*/
func run(port int) {
	//Attempt to set up the server
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		fmt.Println("Couldn't bind socket")
		os.Exit(1)
	}

	fmt.Println("Backend server running at port " + strconv.Itoa(port))

	//The main loop of the backend server
	for {
		//Accept a connection
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprint(os.Stderr, "Failed to accept")
			os.Exit(1)
		}

		defer conn.Close()
		fmt.Fprintln(os.Stderr, "Accepted connection from", conn.RemoteAddr())

		//Read a message from the connection
		buf := make([]byte, 1024)
		reqLen, bufErr := conn.Read(buf)
		if bufErr != nil {
			fmt.Println("Error reading:", err.Error())
		}
		message := FrontendMessage{}
		json.Unmarshal(buf[:reqLen], &message)
		fmt.Println("Message received:" + string(buf))

		//Handle the message and respond
		response := handleMessage(message)
		res, _ := json.Marshal(response)
		conn.Write([]byte(res))

		fmt.Fprintln(os.Stderr, "connection ended")
	}
}

/*
This function processes an incoming message from the frontend server and respond accordingly
Input: a FrontendMessage object
Output: a BackendMessage object
*/
func handleMessage(message FrontendMessage) BackendMessage {
	switch message.Operation {
	case "home":
		return BackendMessage{Success: true, Homeworks: homeworks}
	case "getOne":
		homework := homeworks[message.ID]
		return BackendMessage{Success: true, Homework: homework}
	case "edit":
		id := message.ID
		homeworks[id] = message.NewHomework
		return BackendMessage{Success: true}
	case "create":
		homeworks = append(homeworks, message.NewHomework)
		return BackendMessage{Success: true}
	case "delete":
		id := message.ID
		homeworks = append(homeworks[:id], homeworks[id+1:]...)
		return BackendMessage{Success: true}
	default:
		return BackendMessage{Success: false, ErrMessage: "Unknown error"}
	}
}
