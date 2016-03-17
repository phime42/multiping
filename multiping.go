package main


import (
	"os/exec"
	"os"
	"net"
	"time"
	"log"
	"fmt"
)

type Host struct {
	hostname string
	online bool
	latency time.Duration
}

func hostOnline(address string) (bool) {

	parsed_address := net.ParseIP(address)
	tool := ""
	timeout := ""

	if parsed_address.To16() == nil {
		// the given string is no IP address
		return true
	}

	if parsed_address.To4() != nil {
		// address is v4 parsable, so no v6
		tool = "ping"
		timeout = "-W2"
	} else {
		tool = "ping6"
		timeout = "-i2"
	}

	// ping the host
	_, err := exec.Command(tool,"-c1",timeout,parsed_address.To16().String() ).Output()
	// if host is dead, return 1
	if err != nil {
		return false
	}
	// otherwise, return 0
	return true

}

func check(e error) {
	if e != nil {
			panic(e)
	}
}

func logStatus() {
	filename:="log.txt"
	exec.Command("touch",filename) //error handling would be nice
	logFile, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)

	check(err)
	defer logFile.Close()
	if err != nil {
		panic(err)
  }
	defer logFile.Close()
	// direct all log messages to log.txt
	log.SetOutput(logFile)

	channel := make(chan Host, 3)

	go func() {
		start := time.Now()
		online := hostOnline("8.8.8.8")
		channel <- Host{"Google DNS", online,time.Since(start)}
	}()
	go func() {
		start := time.Now()
		online := hostOnline("127.0.0.1")
		channel <- Host{"Host v4", online,time.Since(start)}
	}()
	go func() {
		start := time.Now()
		online := hostOnline("::1")
		channel <- Host{"Host v6", online, time.Since(start)}
	}()
	go func() {
		// poison pill, closes the buffer after 2000ms allowing for futher exec
		time.Sleep(2500 * time.Millisecond)
		close(channel)
	}()

	message := ""
	for elem := range channel {
		message += fmt.Sprintf("// %v: %v %v", elem.hostname, elem.online, elem.latency)
	}
	log.Printf(message)
	logFile.Sync()
}


func main() {
	for {
		go func() {
			logStatus()
		}()
		time.Sleep(2000 * time.Millisecond)
	}
}
