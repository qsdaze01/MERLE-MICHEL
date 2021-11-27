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

var RCVSIZE = 1024

var lockMessageSent []sync.Mutex
var lockBufferMessage []sync.Mutex
var lockInstanciation sync.Mutex

var bufferMessage [][][]byte
var messageSent []int

var window = 10

func readAckMessage(messageAck []byte) int {
	numAck := string(messageAck[3:9])
	res, _ := strconv.Atoi(numAck)
	return res
}

func receiveAck(clientAddress *net.UDPAddr, socketCommunication *net.UDPConn, RTT float64, instanceNumber int, numSeqReceived int) (int, []byte, int) {
	messageAck := make([]byte, RCVSIZE)
	if RTT == 0 {
		_ = socketCommunication.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		_, _, err := socketCommunication.ReadFromUDP(messageAck)
		if err != nil {
			fmt.Println("Renvoi message")
			fmt.Println(err)

			messageResent := make([]byte, RCVSIZE)
			lockBufferMessage[instanceNumber].Lock()
			if len(bufferMessage[instanceNumber]) >= window+1 {
				messageResent = bufferMessage[instanceNumber][window]
			} else {
				messageResent = bufferMessage[instanceNumber][0]
			}
			lockBufferMessage[instanceNumber].Unlock()
			fmt.Println(string(messageResent))
			_, err := socketCommunication.WriteToUDP(messageResent, clientAddress)
			if err != nil {
				fmt.Println(err)
				return -1, nil, numSeqReceived
			}

			_, messageAck, numSeqReceived = receiveAck(clientAddress, socketCommunication, RTT, instanceNumber, numSeqReceived)
		}
	} else {
		_ = socketCommunication.SetReadDeadline(time.Now().Add(time.Duration(RTT)))
		_, _, err := socketCommunication.ReadFromUDP(messageAck)
		if err != nil {
			fmt.Println("Renvoi message")
			fmt.Println(err)

			messageResent := make([]byte, RCVSIZE)
			lockBufferMessage[instanceNumber].Lock()
			if len(bufferMessage[instanceNumber]) >= window+1 {
				messageResent = bufferMessage[instanceNumber][window]
			} else {
				messageResent = bufferMessage[instanceNumber][0]
			}
			lockBufferMessage[instanceNumber].Unlock()
			_, err := socketCommunication.WriteToUDP(messageResent, clientAddress)
			if err != nil {
				fmt.Println(err)
				return -1, nil, numSeqReceived
			}

			_, messageAck, numSeqReceived = receiveAck(clientAddress, socketCommunication, RTT, instanceNumber, numSeqReceived)
		}
	}
	numSeqReceived = readAckMessage(messageAck)

	return 0, messageAck, numSeqReceived
}

func receive(clientAddress *net.UDPAddr, socketCommunication *net.UDPConn, RTT float64, instanceNumber int) (int, []byte) {
	numSeqReceived := 1
	checkNumSeq := 1
	for {
		for {
			lockBufferMessage[instanceNumber].Lock()
			lengthBufferMessage := len(bufferMessage[instanceNumber])
			lockBufferMessage[instanceNumber].Unlock()
			if lengthBufferMessage == 0 {
				break
			}

			for {
				_, _, numSeqReceived = receiveAck(clientAddress, socketCommunication, RTT, instanceNumber, numSeqReceived)
				if numSeqReceived >= checkNumSeq {
					fmt.Print("checkNumSeq : ")
					fmt.Println(checkNumSeq)
					checkNumSeq++
					break
				}
			}

			lockBufferMessage[instanceNumber].Lock()
			bufferMessage[instanceNumber] = append(bufferMessage[instanceNumber][:0], bufferMessage[instanceNumber][1:]...)
			lockBufferMessage[instanceNumber].Unlock()

			lockMessageSent[instanceNumber].Lock()
			messageSent[instanceNumber]--
			lockMessageSent[instanceNumber].Unlock()
		}
	}
}

func send(clientAddress *net.UDPAddr, socketCommunication *net.UDPConn, window int, instanceNumber int, file *os.File) int {
	seq := []byte("000001")
	numSeq, err := strconv.Atoi(string(seq))

	if err != nil {
		fmt.Println(err)
		return -1
	}

	for {
		lockMessageSent[instanceNumber].Lock()
		packetCount := messageSent[instanceNumber]
		lockMessageSent[instanceNumber].Unlock()

		for packetCount < window {
			fileBuffer := make([]byte, RCVSIZE-6)
			_, errEof := file.Read(fileBuffer)
			if errEof == io.EOF {
				eof := make([]byte, 3)
				eof[0] = byte('F')
				eof[1] = byte('I')
				eof[2] = byte('N')
				_, err := socketCommunication.WriteToUDP(eof, clientAddress)
				if err != nil {
					fmt.Println(err)
					return -1
				}

				/*endTimer := time.Now()
				diffTimer := endTimer.Sub(startTimer)
				fmt.Println(diffTimer)*/

				return 0 //On n'acquitte pas le FIN. C'est normal à cause du break
			} else if errEof != nil {
				fmt.Println(errEof)
				return -1
			} else {
				message := append(seq, fileBuffer...)

				lockBufferMessage[instanceNumber].Lock()
				bufferMessage[instanceNumber] = append(bufferMessage[instanceNumber], message)
				lockBufferMessage[instanceNumber].Unlock()

				_, err := socketCommunication.WriteToUDP(message, clientAddress)
				if err != nil {
					fmt.Println(err)
					return -1
				}
			}

			packetCount++

			lockMessageSent[instanceNumber].Lock()
			messageSent[instanceNumber]++
			lockMessageSent[instanceNumber].Unlock()

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
}

func communicate(wg *sync.WaitGroup, port string, instanceNumber int) {
	var lockMessage sync.Mutex
	var lockBuffer sync.Mutex

	message := make([][]byte, window)

	lockInstanciation.Lock()
	lockMessageSent = append(lockMessageSent, lockMessage)
	lockBufferMessage = append(lockBufferMessage, lockBuffer)
	bufferMessage = append(bufferMessage, message)
	messageSent = append(messageSent, 0)
	lockInstanciation.Unlock()

	RTT := 0.0

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

	go send(clientAddress, socketCommunication, window, instanceNumber, file)

	go receive(clientAddress, socketCommunication, RTT, instanceNumber)

}

func main() {
	instanceNumber := 0
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
			go communicate(&wg, strconv.Itoa(portCommunication), instanceNumber)
			instanceNumber++

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
