package main

import (
	"io"
	"net"
	"time"

	"log"

	"os"
)

var (
	log_file *os.File
	logger   log.Logger
)

func init() {
	file, err := os.OpenFile("sensor.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}

	log_file = file

	logger = *log.New(io.MultiWriter(os.Stdout, log_file), "sensor ", 0)
}

func cleanup() {

}

func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	r[0] = byte(val & 0xFF000000)
	r[1] = byte(val & 0x00FF0000)
	r[2] = byte(val & 0x0000FF00)
	r[3] = byte(val & 0x000000FF)
	return r
}

func handle(err error) {
	if err != nil {
		logger.Panicln(err)
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:1337")
	handle(err)

	defer listener.Close()

	sensor := NewMockSensor("sensor")
	collector := NewCollectorManager(NewExecutor(), "game")
loop:
	for {
		game, err := listener.Accept()
		handle(err)

		logger.Println("got connection")
		sensor.Start()
		logger.Println("started sensor")

		handle(collector.Setup())
		handle(collector.Launch())

		logger.Println("started collector")

		ticker := time.NewTicker(10 * time.Millisecond)

		for {
			select {
			case <-sensor.LiveProcesses():
				_, err := game.Write([]byte{0x01})
				if err != nil {
					logger.Println("exiting...")
					break loop
				}
			case <-sensor.LiveConnections():
				_, err := game.Write([]byte{0x02})
				if err != nil {
					logger.Println("exiting...")
					break loop
				}
			case <-ticker.C:
				_, err := game.Write([]byte{0x00})
				if err != nil {
					logger.Println("exitting...")
					break loop
				}
			}
		}
	}

	logger.Println("Tearing down collector...")
	handle(collector.TearDown())
}
