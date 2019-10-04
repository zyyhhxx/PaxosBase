package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

//Data structure for the data
type Homework struct {
	Name        string
	Desc        string
	Submissions int
}

type FrontendMessage struct {
	Operation   string
	ID          int
	NewHomework Homework
}

type BackendMessage struct {
	Success    bool
	Homeworks  []HomeworkStore
	ErrMessage string
	Homework   Homework
}

type HomeworkStore struct {
	Homework Homework
	Deleted  bool
}

//Global variable for the data
var homeworks = make([]HomeworkStore, 0, 5)
var hwlocks = make([]sync.RWMutex, 0, 5)
var counter = 0

var databaseLock sync.RWMutex

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
	addHomework(&Homework{"hw1", "Getting to know Go", 0})
	addHomework(&Homework{"proj1", "Step 1 to the grand project", 0})
	addHomework(&Homework{"hw2", "Getting to know Go again", 0})
	addHomework(&Homework{"proj2", "Step 2 to the grand project", 0})
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
		go handleConnection(conn, err)
	}
}

/*
This function handles the request for the home page
Input: an iris context
Output: none
*/
func handleConnection(conn net.Conn, err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to accept")
		os.Exit(1)
	}

	defer conn.Close()
	//fmt.Fprintln(os.Stderr, "Accepted connection from", conn.RemoteAddr())

	//Read a message from the connection
	buf := make([]byte, 1024)
	reqLen, bufErr := conn.Read(buf)
	if bufErr != nil {
		fmt.Println("Error reading:", bufErr.Error())
	}
	message := FrontendMessage{}
	json.Unmarshal(buf[:reqLen], &message)
	//fmt.Println("Message received:" + string(buf))

	//Handle the message and respond
	response := handleMessage(message)
	res, _ := json.Marshal(response)
	conn.Write([]byte(res))

	//fmt.Fprintln(os.Stderr, "connection ended")
}

/*
This function processes an incoming message from the frontend server and respond accordingly
Input: a FrontendMessage object
Output: a BackendMessage object
*/
func handleMessage(message FrontendMessage) BackendMessage {
	switch message.Operation {
	case "home":
		return readAll()
	case "getOne":
		return readOne(message.ID)
	case "edit":
		return editHomework(message.ID, &message.NewHomework)
	case "create":
		return addHomework(&message.NewHomework)
	case "delete":
		return removeHomework(message.ID)
	case "ping":
		return BackendMessage{Success: true}
	default:
		return BackendMessage{Success: false, ErrMessage: "Unknown operation"}
	}
}

/*
This function returns the entire list of homework
Input: None
Output: a BackendMessage object
*/
func readAll() BackendMessage {
	databaseLock.RLock()
	defer databaseLock.RUnlock()

	return BackendMessage{Success: true, Homeworks: homeworks}
}

/*
This function returns the homework object specified by the integer
Input: an integer
Output: a BackendMessage object
*/
func readOne(id int) BackendMessage {
	databaseLock.RLock()
	if id >= len(homeworks) {
		return BackendMessage{Success: false, ErrMessage: "Index out of range"}
	}

	hwlocks[id].RLock()
	homework := homeworks[id]
	hwlocks[id].RUnlock()

	databaseLock.RUnlock()
	return BackendMessage{Success: true, Homework: homework.Homework}
}

/*
This function creates an entry in the databse for the input Homework object
Input: a pointer of Homework
Output: a BackendMessage object
*/
func addHomework(hw *Homework) BackendMessage {
	hs := HomeworkStore{*hw, false}

	databaseLock.Lock()

	//Add the entry to the database
	homeworks = append(homeworks, hs)

	//Add a corresponding mutex to the mutex list
	var m sync.RWMutex
	hwlocks = append(hwlocks, m)

	databaseLock.Unlock()
	return BackendMessage{Success: true}
}

/*
This function edits the entry in the database specified by the integer based on the input Homework object
Input: a pointer of Homework and an integer
Output: a BackendMessage object
*/
func editHomework(id int, hw *Homework) BackendMessage {
	//Even though this function writes to the database, it only acquires the database read lock,
	//because multiple edit operations need to be allowed as long as they are not editing the same entry,
	//but edit operations still need to block add operations, which might cause a resizing of the slice
	databaseLock.RLock()
	defer databaseLock.RUnlock()

	if id >= len(homeworks) {
		return BackendMessage{Success: false, ErrMessage: "Index out of range"}
	}

	hwlocks[id].Lock()
	defer hwlocks[id].Unlock()

	if !homeworks[id].Deleted {
		homeworks[id].Homework.Desc = hw.Desc
		homeworks[id].Homework.Name = hw.Name
		homeworks[id].Homework.Submissions++

		return BackendMessage{Success: true}
	} else {
		return BackendMessage{Success: false, ErrMessage: "Specified entry is deleted. Cannot edit. "}
	}
}

/*
This function marks the entry in the database specified by the integer as deleted
Input: an integer
Output: a BackendMessage object
*/
func removeHomework(id int) BackendMessage {
	databaseLock.RLock()
	defer databaseLock.RUnlock()

	if id >= len(homeworks) {
		return BackendMessage{Success: false, ErrMessage: "Index out of range"}
	}

	hwlocks[id].Lock()
	defer hwlocks[id].Unlock()

	if !homeworks[id].Deleted {
		homeworks[id].Deleted = true
		return BackendMessage{Success: true}
	} else {
		return BackendMessage{Success: false, ErrMessage: "Specified entry is already deleted. Cannot delete again. "}
	}
}
