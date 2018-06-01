package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
)

var (
	ifaceName  = flag.String("i", "eth0", "Interface name")
	outputFile = flag.String("o", "", "Output file path or STDOUT")
)

func main() {
	flag.Parse()

	outChan := make(chan string)
	go func() {
		log.Fatal(listenDHCPPacket(*ifaceName, outChan))
	}()

	var f *os.File
	var err error

	if *outputFile != "" {
		f, err = os.OpenFile(*outputFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}

	var lastTime time.Time
	var timeM sync.Mutex
	clearTimer := time.NewTicker(time.Second * 5)
	defer clearTimer.Stop()

	writeBuf := make(chan string)

	go func() {
		for {
			select {
			case <-clearTimer.C:
				timeM.Lock()
				t := lastTime
				timeM.Unlock()
				if time.Now().Sub(t) > time.Second*10 {
					writeBuf <- "Waiting..."
				}
			}
		}
	}()

	go func() {
		for out := range outChan {
			timeM.Lock()
			lastTime = time.Now()
			timeM.Unlock()
			writeBuf <- out
		}
	}()

	for out := range writeBuf {
		if f != nil {
			f.Truncate(0)
			f.WriteAt([]byte(out), 0)
		} else {
			fmt.Fprintf(os.Stdout, "\r%s", out)
		}
	}
}

func listenDHCPPacket(ifaceName string, output chan<- string) error {
	output <- "waiting interface: " + ifaceName
	conn, err := conn.NewUDP4BoundListener(ifaceName, ":67")
	if err != nil {
		return err
	}
	defer conn.Close()

	buffer := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		if n < 240 {
			continue
		}
		req := dhcp4.Packet(buffer[:n])
		if req.HLen() > 16 {
			continue
		}
		go func() { output <- req.CHAddr().String() }()
	}
}
