package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
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

	//Run the server
	app.Run(iris.Addr(":" + strconv.Itoa(port)))
}

/*
This function handles the request for the home page
Input: an iris context
Output: none
*/
func handleHome(ctx iris.Context) {

	response := sendToBackend(FrontendMessage{Operation: "home"})
	if !response.Success {
		ctx.HTML("<h1>Error loading the home page</h1>")
	}

	homeworks := response.Homeworks

	ctx.HTML("<h1>List of homeworks</h1>")
	if len(homeworks) > 0 {
		ctx.HTML("<table>")
		ctx.HTML("<tr><th>Homework</th><th>Description</th></tr>")
		for index, element := range homeworks {
			ctx.HTML("<tr><td> <a href=\"/edit?id=" + strconv.Itoa(index) +
				"\">" + element.Name + "</a></td><td>" + element.Desc + "</th></td>")
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
		ctx.HTML("<h1>Error getting the specified entry</h1>")
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

	response := sendToBackend(FrontendMessage{Operation: "edit", ID: id, NewHomework: Homework{name, desc}})
	if !response.Success {
		ctx.HTML("<h1>Error editting the specified entry</h1>")
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

	response := sendToBackend(FrontendMessage{Operation: "create", NewHomework: Homework{name, desc}})
	if !response.Success {
		ctx.HTML("<h1>Error creating a new entry</h1>")
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
		ctx.HTML("<h1>Error deleting the specified entry</h1>")
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
		fmt.Fprint(os.Stderr, "could not connect: ", err.Error())
	}
	defer conn.Close()

	//Send the message to the backend server
	mes, _ := json.Marshal(message)
	conn.Write(mes)

	//Receive and return the response
	buf := make([]byte, 1024)
	reqLen, bufErr := conn.Read(buf)
	if bufErr != nil {
		fmt.Println("Error reading:", err.Error())
	}
	response := BackendMessage{}
	json.Unmarshal(buf[:reqLen], &response)
	fmt.Println("Message received:" + string(buf))
	return response
}
