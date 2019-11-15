# PaxosBase
PaxosBase is a scalable distributed database system which supports CRUD operations and hosts a distributed database coordinated by Paxos protocol, allowing a large number of users to create, read, update and delete items in the database. PaxosBase features a frontend built by HTML and a backend built by Go

## How to build: 
`go run frontend.go --listen (custom port number) --backend (custom address of the backend server)`  
`go run backend.go --listen (custom port number)`  
`go run tests.go`
