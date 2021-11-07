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

	for {
		fileBuffer := make([]byte, 32)
		_, errEof := file.Read(fileBuffer)

		_, err := socketCommunication.WriteToUDP(fileBuffer, clientAddress)

		if errEof == io.EOF {
			eof := []byte("EOF")
			_, err := socketCommunication.WriteToUDP(eof, clientAddress)
			if err != nil {
				fmt.Println(err)
				return
			}
			break
		} else if errEof != nil {
			fmt.Println(errEof)
			return
		}

		if err != nil {
			fmt.Println(err)
			return
		}
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
				fmt.Println("Handshaked !")
			}

			portCommunication++
		}
	}

	wg.Wait()

}
