package main

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

type Test struct {
	workers  uint64
	rate     vegeta.Rate
	duration time.Duration
	target   vegeta.Target
	wg       *sync.WaitGroup
}

/*
The main loop of the program
Input: none
Output: none
*/
func main() {
	//This class has 60 students and 3 instructors
	for i := 0; i < 10; i++ {
		fmt.Println("Round " + strconv.Itoa(i))
		readTest()
		createTest()
		updateTest()
		updateCreateTest()
	}
}

/*
This function implements a Read only test
Input: none
Output: none
*/
func readTest() {
	//Students come in and check out the homework
	fmt.Println("Test 1: A simple concurrent read test")
	runTest(Test{
		workers:  60,
		rate:     vegeta.Rate{Freq: 100, Per: time.Second},
		duration: 4 * time.Second,
		target: vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/",
		},
	})
	fmt.Println("-------------------------------------------------")
}

/*
This function implements a Create only test
Input: none
Output: none
*/
func createTest() {
	//Instructors decide to give students more homework
	fmt.Println("Test 2: A simple concurrent create test")
	runTest(Test{
		workers:  3,
		rate:     vegeta.Rate{Freq: 100, Per: time.Second},
		duration: 4 * time.Second,
		target: vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/create_form?itemName=projN&desc=Step+N+of+the+project",
		},
	})
	fmt.Println("-------------------------------------------------")
}

/*
This function implements an Update only test
Input: none
Output: none
*/
func updateTest() {
	//Students start to submit homework
	fmt.Println("Test 3: A simple concurrent update test")
	var wg sync.WaitGroup
	wg.Add(1)
	go runTest(Test{
		20,
		vegeta.Rate{Freq: 100, Per: time.Second},
		2 * time.Second,
		vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/edit_form?id=0&itemName=hw1&desc=Getting+to+know+Go",
		},
		&wg,
	})
	wg.Add(1)
	go runTest(Test{
		20,
		vegeta.Rate{Freq: 100, Per: time.Second},
		2 * time.Second,
		vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/edit_form?id=1&itemName=proj1&desc=Step+1+to+the+grand+project",
		},
		&wg,
	})
	wg.Add(1)
	go runTest(Test{
		20,
		vegeta.Rate{Freq: 100, Per: time.Second},
		2 * time.Second,
		vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/edit_form?id=2&itemName=hw2&desc=Getting+to+know+Go+again",
		},
		&wg,
	})
	wg.Wait()
	fmt.Println("-------------------------------------------------")
}

/*
This function implements a mixed Create-Update test
Input: none
Output: none
*/
func updateCreateTest() {
	//Students are submiting homework while intstructors are creating even more homework
	fmt.Println("Test 4: A simple concurrent update and create test")
	var wg sync.WaitGroup
	wg.Add(1)
	go runTest(Test{
		20,
		vegeta.Rate{Freq: 100, Per: time.Second},
		2 * time.Second,
		vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/edit_form?id=0&itemName=hw1&desc=Getting+to+know+Go",
		},
		&wg,
	})
	wg.Add(1)
	go runTest(Test{
		20,
		vegeta.Rate{Freq: 100, Per: time.Second},
		2 * time.Second,
		vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/edit_form?id=1&itemName=proj1&desc=Step+1+to+the+grand+project",
		},
		&wg,
	})
	wg.Add(1)
	go runTest(Test{
		20,
		vegeta.Rate{Freq: 100, Per: time.Second},
		2 * time.Second,
		vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/edit_form?id=2&itemName=hw2&desc=Getting+to+know+Go+again",
		},
		&wg,
	})
	wg.Add(1)
	go runTest(Test{
		workers:  3,
		rate:     vegeta.Rate{Freq: 10, Per: time.Second},
		duration: 2 * time.Second,
		target: vegeta.Target{
			Method: "GET",
			URL:    "http://localhost:8080/create_form?itemName=projN&desc=Step+N+of+the+project",
		},
		wg: &wg,
	})
	wg.Wait()
	fmt.Println("-------------------------------------------------")
}

/*
This function runs a test based on the information given by the Test struct
Input: a Test object
Output: none
*/
func runTest(test Test) {
	if test.wg != nil {
		defer test.wg.Done()
	}

	targeter := vegeta.NewStaticTargeter(test.target)
	attacker := vegeta.NewAttacker(vegeta.Workers(test.workers))

	var metrics vegeta.Metrics

	for res := range attacker.Attack(targeter, test.rate, test.duration, "") {
		metrics.Add(res)
	}

	metrics.Close()

	var buf bytes.Buffer
	vegeta.NewTextReporter(&metrics)(&buf)
	fmt.Println(buf.String())
}
