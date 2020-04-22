package main

import (
	"fmt"
	"go.bug.st/serial.v1"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var min, max = 50, 70
var cycles = 0
var commands []string

func start(w http.ResponseWriter, r *http.Request) {
	cycles, _ = strconv.Atoi(r.FormValue("cycle"))
	createCode()                                                                  //first, create the g-code
	resp := "<script type='text/javascript'> window.location.href = '/'</script>" //redirect the user back to their usual screen
	_, err := fmt.Fprintf(w, resp)
	if err != nil {
		log.Fatal(err)
	}
}

func exit(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(w, "Good Bye...")
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/start", start)
	http.HandleFunc("/exit", exit)
	fmt.Println("Server is running on port 8000!")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createCode() {
	for i := 0; i <= cycles; i++ {
		diameter := rand.Intn(max-min+1) + min //this gives us a random number within our borders

		//Absolute Borders are X260, Y630 Create new virtual borders here
		xMax := 260 - diameter
		yMax := 630 - diameter

		//Define a starting point
		startX := rand.Intn(xMax-diameter+1) + diameter
		startY := rand.Intn(yMax-diameter+1) + diameter

		Mx, My := 0, 0 //create the variables we need for the circle

		dir := rand.Intn(4) //select a random direction

		switch dir {
		case 0: //Down
			Mx = -diameter / 2
			My = 0
		case 1: //Up
			Mx = diameter / 2
			My = 0
		case 2: //Left
			Mx = 0
			My = -diameter / 2
		case 3: //Right
			Mx = 0
			My = diameter / 2
		default:
			log.Fatal("ERROR: Wrong direction!")
		}

		//create the g-code
		cmd1 := "G01 X" + strconv.Itoa(startX) + " Y" + strconv.Itoa(startY) + "\n"
		cmd2 := "G02 X" + strconv.Itoa(startX) + " Y" + strconv.Itoa(startY) + " I" + strconv.Itoa(Mx) + " J" + strconv.Itoa(My) + "\n"

		//add the g-code to a slice
		commands = append(commands, cmd1, cmd2)

	}
	send() //now send the g-code to grbl
}

func send() {
	mode := &serial.Mode{ //settings for serial port
		BaudRate: 9600, //TODO: change to grbl baud rate
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	var port, err = serial.Open("COM6", mode) //TODO: Pass in correct port name
	if err != nil {
		log.Fatal(err)
	}

	//read the welcome msg from grbl
	buff := make([]byte, 100) //a buffer to store the msg
	for i := 0; i < len(buff); i++ {
		// Reads up to 100 bytes
		n, err := port.Read(buff)
		if err != nil {
			log.Fatal(err)

		}
		if n == 0 { //For some reason, this wont work...
			fmt.Println("\nEOF")

			break
		}

		//print out the msg we received
		msg := string(buff[:n])
		fmt.Printf("%v", msg)

		if i >= 9 { //TODO: Check what number to use for GRBL, it might work with this number because of the delay
			break
		}
	}
	fmt.Println("Done")
	time.Sleep(time.Second) //Wait a second so we dont kill the Port

	bytes := 0 //we use this to count the overall bytes sent

	//write the commands
	for i := 0; i < len(commands); i++ { //TODO: Check if this is a good way of sending or if we are going to fast
		n, err := port.Write([]byte(commands[i]))
		if err != nil {
			log.Fatal(err)
		}
		bytes += n
		//time.Sleep(time.Second) we might need this to slow down sending...
	}
	fmt.Printf("Wrote %b bytes\n", bytes)
	err2 := port.Close()

	if err2 != nil {
		log.Fatal(err2)
	}
}
