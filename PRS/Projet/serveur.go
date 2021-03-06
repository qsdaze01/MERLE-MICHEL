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

func receiveACK(socketCommunication net.UDPConn, messageAck []byte, RTT float64) (int, []byte) {
	if RTT == 0 {
		err := socketCommunication.SetReadDeadline(time.Now().Add(40 * time.Millisecond))
		if err != nil {
			fmt.Println(err)
			return -1, nil
		}
	} else {
		err := socketCommunication.SetReadDeadline(time.Now().Add(time.Duration(RTT)))
		if err != nil {
			fmt.Println(err)
			return -1, nil
		}
	}

	_, _, err := socketCommunication.ReadFromUDP(messageAck)

	if err != nil {
		fmt.Println(err)
		return -1, nil
	}

	return 0, messageAck
}

func send(message []byte, clientAddress *net.UDPAddr, socketCommunication net.UDPConn) int {

	_, err := socketCommunication.WriteToUDP(message, clientAddress)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return 0
}

func communicate(wg *sync.WaitGroup, port string) {

	RCVSIZE := 1024
	RTT := 0.0
	window := 1
	messageSent := 0
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

	filenameClient := make([]byte, 1024)
	lengthFilenameClient, clientAddress, err := socketCommunication.ReadFromUDP(filenameClient)
	fmt.Println(string(filenameClient[0:lengthFilenameClient]))

	file, err := os.Open(string(filenameClient[0 : lengthFilenameClient-1]))

	if err != nil {
		fmt.Println(err)
		return
	}

	seq := []byte("000001")
	numSeq, err := strconv.Atoi(string(seq))

	if err != nil {
		fmt.Println(err)
		return
	}

	var bufferMessage [][]byte
	startTimer := time.Now()

	if messageSent < window {

	}
	for {
		messageAck := make([]byte, RCVSIZE)
		difftime := 100 * time.Millisecond
		res := 1
		begin := time.Now()

		for i := messageSent; i < window; i++ {
			fileBuffer := make([]byte, RCVSIZE-6)
			_, errEof := file.Read(fileBuffer)
			message := append(seq, fileBuffer...)

			if errEof == io.EOF {
				eof := make([]byte, 3)
				eof[0] = byte('F')
				eof[1] = byte('I')
				eof[2] = byte('N')
				_, err := socketCommunication.WriteToUDP(eof, clientAddress)
				if err != nil {
					fmt.Println(err)
					return
				}

				endTimer := time.Now()
				diffTimer := endTimer.Sub(startTimer)
				fmt.Println(diffTimer)

				break //On n'acquitte pas le FIN. C'est normal ?? cause du break
			} else if errEof != nil {
				fmt.Println(errEof)
				return
			}
			begin = time.Now()
			res = send(message, clientAddress, *socketCommunication)
			if res < 0 {
				fmt.Println("Probl??me d'envoi")
				return
			}
			bufferMessage = append(bufferMessage, message)
			messageSent++
		}

		for {
			res, messageAck = receiveACK(*socketCommunication, messageAck, RTT)
			if res != 0 {
				window = 1
				RTT = 4 * float64(difftime)
				res = send(bufferMessage[0], clientAddress, *socketCommunication)
				if res < 0 {
					fmt.Println("Probl??me d'envoi")
					return
				}
			} else {
				window++
				messageSent--
				RTT = RTT - 0.1*(RTT-4*float64(difftime))
				bufferMessage = append(bufferMessage[:0], bufferMessage[1:]...)
				break
			}
		}

		end := time.Now()

		difftime = end.Sub(begin)
		//fmt.Println(difftime)

		fmt.Println("Renvoi message")

		fmt.Println(numSeq)
		numSeq++
		if numSeq < 10 {
			seq = []byte("00000" + strconv.Itoa(numSeq))
		} else if numSeq < 100 {
			seq = []byte("0000" + strconv.Itoa(numSeq))
		} else if numSeq < 1000 {
			seq = []byte("000" + strconv.Itoa(numSeq))
		} else if numSeq < 10000 {
			seq = []byte("00" + strconv.Itoa(numSeq))
		} else if numSeq < 100000 {
			seq = []byte("0" + strconv.Itoa(numSeq))
		} else {
			seq = []byte(strconv.Itoa(numSeq))
		}

	}
}

func main() {
	rand.Seed(time.Now().Unix())

	var wg sync.WaitGroup

	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("args : port")
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

	portCommunication := 2000
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		handshake1 := make([]byte, 1024)
		_, clientAddress, err := socketConnect.ReadFromUDP(handshake1)
		if string(handshake1[0:3]) == "SYN" {
			handshake2 := "SYN-ACK" + strconv.Itoa(portCommunication)
			fmt.Println(handshake2)
			_, err = socketConnect.WriteToUDP([]byte(handshake2), clientAddress)

			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("SYN-ACK")

			//go routine avec socket communication ici
			wg.Add(1)
			go communicate(&wg, strconv.Itoa(portCommunication))

			handshake3 := make([]byte, 1024)
			lengthHandshake3, _, err := socketConnect.ReadFromUDP(handshake3)

			if err != nil {
				fmt.Println(err)
				return
			}

			if string(handshake3[0:lengthHandshake3-1]) == "ACK" {
				fmt.Println("Handshaked !")
			}

			portCommunication++
		}
	}
}
