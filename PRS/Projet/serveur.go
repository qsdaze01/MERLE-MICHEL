package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func communicate(wg *sync.WaitGroup, port string) {
	RCVSIZE := 42
	RTT := 0.0
	timeout := 0
	defer wg.Done()

	portCommunication := ":" + port
	communicationParameters, err := net.ResolveUDPAddr("udp4", portCommunication)

	if err != nil {
		fmt.Println(err)
		return
	}

	socketCommunication, err := net.ListenUDP("udp4", communicationParameters)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer socketCommunication.Close()

	filenameClient := make([]byte, 1024)
	lengthFilenameClient, clientAddress, err := socketCommunication.ReadFromUDP(filenameClient)
	fmt.Println(string(filenameClient[0:lengthFilenameClient]))

	file, err := os.Open(string(filenameClient[0:lengthFilenameClient]))

	if err != nil {
		fmt.Println(err)
		return
	}

	seq := []byte(strconv.Itoa(random(0, 2000000)))
	numSeq, err := strconv.Atoi(string(seq))

	if err != nil {
		fmt.Println(err)
		return
	}
	count := 0
	for {
		//fmt.Println(count)
		count++
		fileBuffer := make([]byte, RCVSIZE-10)
		for i := range fileBuffer {
			fileBuffer[i] = 0
		}
		_, errEof := file.Read(fileBuffer)
		message := append(fileBuffer, seq...)

		if errEof == io.EOF {
			//fmt.Println(count)
			eof := make([]byte, RCVSIZE-10)
			for i := range eof {
				eof[i] = 0
			}
			eof[0] = byte('E')
			eof[1] = byte('O')
			eof[2] = byte('F')
			eofMessage := append(eof, seq...)
			_, err := socketCommunication.WriteToUDP(eofMessage, clientAddress)
			if err != nil {
				fmt.Println(err)
				return
			}
			break //On n'acquitte pas le EOF. C'est normal à cause du break
		} else if errEof != nil {
			fmt.Println(errEof)
			return
		}

		//fmt.Println(message)
		begin := time.Now()
		_, err := socketCommunication.WriteToUDP(message, clientAddress)

		if err != nil {
			fmt.Println(err)
			return
		}

		chanAck := make(chan []byte, RCVSIZE-10)
		messageAck := make([]byte, RCVSIZE-10)

		go func() {
			_, _, err = socketCommunication.ReadFromUDP(messageAck)
			if err != nil {
				fmt.Println(err)
				fmt.Println("a")
				return
			}
			fmt.Println(string(messageAck))
			chanAck <- messageAck
		}()

		if RTT != 0 {
			select {
			case res := <-chanAck:
				fmt.Println(res)
				fmt.Println(messageAck)
			case <-time.After( /*time.Duration(RTT)*/ 1 * time.Second):
				fmt.Println("Timeout")
				return
			}
		}

		end := time.Now()

		difftime := end.Sub(begin)

		if RTT == 0 || timeout == 1 {
			RTT = 4 * float64(difftime)
			timeout = 0
		} else {
			RTT = RTT - 0.1*(RTT-4*float64(difftime))
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(messageAck))

		if string(messageAck[0:3]) != "ACK" {
			fmt.Println("Pas de ACK reçu")
			return
		}

		numSeq++
		seq = []byte(strconv.Itoa(numSeq))
	}

}

func main() {
	rand.Seed(time.Now().Unix())

	var wg sync.WaitGroup

	arguments := os.Args
	if len(arguments) == 2 {
		fmt.Println("args : port_connection port_communication")
		return
	}
	portConnection := ":" + arguments[1]

	connectionParameters, err := net.ResolveUDPAddr("udp4", portConnection)
	if err != nil {
		fmt.Println(err)
		return
	}

	socketConnect, err := net.ListenUDP("udp4", connectionParameters)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer socketConnect.Close()

	portCommunication, err := strconv.Atoi(arguments[2])
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		handshake1 := make([]byte, 1024)
		lengthHandshake1, clientAddress, err := socketConnect.ReadFromUDP(handshake1)

		if string(handshake1[0:lengthHandshake1]) == "SYN" {
			handshake2 := "SYN_ACK " + strconv.Itoa(portCommunication)
			_, err = socketConnect.WriteToUDP([]byte(handshake2), clientAddress)

			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("SYN_ACK")

			//go routine avec socket communication ici
			wg.Add(1)
			go communicate(&wg, strconv.Itoa(portCommunication))

			handshake3 := make([]byte, 1024)
			lengthHandshake3, _, err := socketConnect.ReadFromUDP(handshake3)

			if err != nil {
				fmt.Println(err)
				return
			}

			if string(handshake3[0:lengthHandshake3]) == "ACK" {
				fmt.Println("Handshake !")
			}

			portCommunication++
		}
	}
	wg.Wait()
}
