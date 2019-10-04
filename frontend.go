package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
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

var backend = ""

/*
The main loop of the program
Input: none
Output: none
*/
func main() {

	//Set flags
	portNumPtr := flag.Int("listen", 8080, "The port number the server will listen at")
	backendPtr := flag.String("backend", "localhost:8090", "The address of the backend server")
	flag.Parse()

	//Parse any command-line arguments
	port := *portNumPtr
	backend = *backendPtr

	//Initialize the server
	app := iris.New()
	app.Use(recover.New())
	app.Use(logger.New())

	//Set up handle
	app.Handle("GET", "/", handleHome)
	app.Get("/edit", handleEdit)
	app.Get("/edit_form", handleEditForm)
	app.Get("/create", handleCreate)
	app.Get("/create_form", handleCreateForm)
	app.Get("/delete", handleDelete)

	go ping()

	//Run the server
	app.Run(iris.Addr(":" + strconv.Itoa(port)))
}

/*
This function keeps pinging the backend server and report a failure if it does not get a response
Input: none
Output: none
*/
func ping() {
	//Pingack
	for {
		response := sendToBackend(FrontendMessage{Operation: "ping"})
		if !response.Success {
			fmt.Println("Detected failure on " + backend + " at " + time.Now().UTC().String())
		}
		time.Sleep(10 * time.Second)
	}
}

/*
This function handles the request for the home page
Input: an iris context
Output: none
*/
func handleHome(ctx iris.Context) {

	response := sendToBackend(FrontendMessage{Operation: "home"})
	if !response.Success {
		ctx.HTML("<h1>Error loading the home page: " + response.ErrMessage + "</h1>")
	}

	homeworks := response.Homeworks

	ctx.HTML("<h1>List of homeworks</h1>")
	ctx.HTML("<h1>Total number of homeworks: " + strconv.Itoa(len(homeworks)) + "</h1>")
	if len(homeworks) > 0 {
		ctx.HTML("<table>")
		ctx.HTML("<tr><th>Homework</th><th>Description</th><th>Submissions</th></tr>")
		for index, element := range homeworks {
			if !element.Deleted {
				ctx.HTML("<tr><td> <a href=\"/edit?id=" + strconv.Itoa(index) +
					"\">" + element.Homework.Name + "</a></td><td>" + element.Homework.Desc + "</td><td>" + strconv.Itoa(element.Homework.Submissions) + "</td></tr>")
			}
		}
		ctx.HTML("</table>")
	} else {
		ctx.HTML("Empty")
	}
	ctx.HTML("<a href=\"/create\">Create a new homework</a>")
}

/*
This function handles the request for the edit page
Input: an iris context
Output: none
*/
func handleEdit(ctx iris.Context) {
	id, _ := strconv.Atoi(ctx.URLParam("id"))

	response := sendToBackend(FrontendMessage{Operation: "getOne", ID: id})
	if !response.Success {
		ctx.HTML("<h1>Error getting the specified entry: " + response.ErrMessage + "</h1>")
	}
	homework := response.Homework

	ctx.HTML("<h1>Edit Homework</h1>")
	ctx.HTML("<form action=\"/edit_form\">")
	ctx.HTML("<input type=\"hidden\" name=\"id\" value=\"" + strconv.Itoa(id) + "\">Name:<br>")
	ctx.HTML("<input type=\"text\" name=\"itemName\" value=\"" + homework.Name + "\"><br>")
	ctx.HTML("Description:<br>")
	ctx.HTML("<input type=\"text\" name=\"desc\" value=\"" + homework.Desc + "\"><br>")
	ctx.HTML("<input type=\"submit\" value=\"Submit\"></form>")
	ctx.HTML("<a href=\"/delete\">Delete homework</a><br>")
	ctx.HTML("<a href=\"/\">Back to home</a>")
}

/*
This function handles the request for the edit_form page
Input: an iris context
Output: none
*/
func handleEditForm(ctx iris.Context) {
	id, _ := strconv.Atoi(ctx.FormValue("id"))
	name := ctx.FormValue("itemName")
	desc := ctx.FormValue("desc")

	response := sendToBackend(FrontendMessage{Operation: "edit", ID: id, NewHomework: Homework{Name: name, Desc: desc}})
	if !response.Success {
		ctx.HTML("<h1>Error editing the specified entry: " + response.ErrMessage + "</h1>")
	} else {
		ctx.HTML("<h1>Homework Updated!</h1>")
	}
	ctx.HTML("<a href=\"/\">Back to home</a>")
}

/*
This function handles the request for the create page
Input: an iris context
Output: none
*/
func handleCreate(ctx iris.Context) {
	ctx.HTML("<h1>Edit Homework</h1>")
	ctx.HTML("<form action=\"/create_form\">Name:<br>")
	ctx.HTML("<input type=\"text\" name=\"itemName\"><br>")
	ctx.HTML("Description:<br><input type=\"text\" name=\"desc\"><br>")
	ctx.HTML("<input type=\"submit\" value=\"Submit\"></form>")
	ctx.HTML("<a href=\"/\">Back to home</a>")
}

/*
This function handles the request for the create_form page
Input: an iris context
Output: none
*/
func handleCreateForm(ctx iris.Context) {
	name := ctx.FormValue("itemName")
	desc := ctx.FormValue("desc")

	response := sendToBackend(FrontendMessage{Operation: "create", NewHomework: Homework{name, desc, 0}})
	if !response.Success {
		ctx.HTML("<h1>Error creating a new entry: " + response.ErrMessage + "</h1>")
	} else {
		ctx.HTML("<h1>Homework Created!</h1>")
	}

	ctx.HTML("<a href=\"/\">Back to home</a>")
}

/*
This function handles the request for the delete page
Input: an iris context
Output: none
*/
func handleDelete(ctx iris.Context) {
	id, _ := strconv.Atoi(ctx.FormValue("id"))

	response := sendToBackend(FrontendMessage{Operation: "delete", ID: id})
	if !response.Success {
		ctx.HTML("<h1>Error deleting the specified entry: " + response.ErrMessage + "</h1>")
	} else {
		ctx.HTML("<h1>Homework Deleted!</h1>")
	}

	ctx.HTML("<a href=\"/\">Back to home</a>")
}

/*
This function sends a message to the backend server and returns the response
Input: a FrontendMessage object
Output: a BackendMessage object
*/
func sendToBackend(message FrontendMessage) BackendMessage {
	//Attempt to connect to the backend server
	conn, err := net.Dial("tcp", backend)
	if err != nil {
		return BackendMessage{Success: false, ErrMessage: err.Error()}
	}
	defer conn.Close()

	//Send the message to the backend server
	mes, _ := json.Marshal(message)
	conn.Write(mes)

	//Set up a timeout
	timeoutDuration := 5 * time.Second
	conn.SetReadDeadline(time.Now().Add(timeoutDuration))

	//Receive and return the response
	buf := make([]byte, 1024*1024) //1MB buffer
	reqLen, bufErr := conn.Read(buf)
	if bufErr != nil {
		fmt.Println("Error reading:", bufErr.Error())
		return BackendMessage{Success: false, ErrMessage: bufErr.Error()}
	}
	response := BackendMessage{}
	json.Unmarshal(buf[:reqLen], &response)
	//fmt.Println("Message received:" + string(buf))
	return response
}
