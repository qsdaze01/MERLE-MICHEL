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

var arg = os.Args
var RCVSIZE, _ = strconv.Atoi(arg[3])
var TIMEOUT int64 = 40000000

var window, _ = strconv.Atoi(arg[2])

//go routine permettant de recevoir en permanence les ack venant du client et de les envoyer à la go routine send pour qu'elle puisse gérer les retransmissions
func receive(channelAck chan int, socketCommunication *net.UDPConn, channelStop chan bool) {
	for {
		select {
		case stop := <-channelStop:
			if stop == true {
				return //on stoppe la go routine
			}
		default:
		}
		messageAck := make([]byte, 10)
		_, _, _ = socketCommunication.ReadFromUDP(messageAck)
		numAck := string(messageAck[3:9])
		res, _ := strconv.Atoi(numAck)
		//fmt.Print("ACK : ")
		//fmt.Println(numAck)
		channelAck <- res
	}
}

func send(clientAddress *net.UDPAddr, socketCommunication *net.UDPConn, file *os.File, channelAck chan int, channelStop chan bool) int {
	seq := []byte("000001")
	numSeq, _ := strconv.Atoi(string(seq))
	var messageSentBuffer [][]byte
	packetCount := 0
	startTimer := time.Now()

	for {

		for packetCount < window {
			fileBuffer := make([]byte, RCVSIZE-6)
			_, errEof := file.Read(fileBuffer)
			if errEof == io.EOF {

				//_, err := socketCommunication.WriteToUDP(eof, clientAddress)

				//fmt.Println("EOF prêt à être envoyé")

			} else if errEof != nil {
				fmt.Println(errEof)
				return -1
			} else {
				message := append(seq, fileBuffer...)
				timestamp := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))
				messageTimestamped := append(timestamp, message...)
				messageSentBuffer = append(messageSentBuffer, messageTimestamped)

				_, err := socketCommunication.WriteToUDP(message, clientAddress)
				if err != nil {
					fmt.Println(err)
					return -1
				}
			}

			packetCount++

			//fmt.Print("Paquet envoyé : ")
			//fmt.Println(numSeq)
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

			select {
			case numAckReceived := <-channelAck:
				for i := 0; i < len(messageSentBuffer); i++ {
					//étape 1: on check les numACK dès qu'on trouve celui qui corrrespond, on supprime le message du buffer, il ne sert plus à rien
					extractNumAck := messageSentBuffer[i][19:25]
					intNumAck, _ := strconv.Atoi(string(extractNumAck))
					if intNumAck <= numAckReceived {
						//fmt.Print("Suppression du buffer : ")
						//fmt.Println(intNumAck)
						messageSentBuffer = append(messageSentBuffer[:i], messageSentBuffer[i+1:]...) //on retire le message qui a été acquitté

						packetCount--
					}
				}
			default:
			}
		}

		if len(messageSentBuffer) == 0 {
			eof := make([]byte, 3)
			eof[0] = byte('F')
			eof[1] = byte('I')
			eof[2] = byte('N')
			_, _ = socketCommunication.WriteToUDP(eof, clientAddress)
			endTimer := time.Now()
			diffTimer := endTimer.Sub(startTimer)
			//fmt.Println("EOF envoyé, fichier transféré avec succès !")
			fmt.Println(diffTimer)
			channelStop <- true //on dit à la go routine receive de s'arrêter
			return 0            //on s'arrête quand on a tout reçu
		}

		for i := 0; i < len(messageSentBuffer); i++ {
			//étape 2: on check les timestamp et on réémet les messages qui sont en timeout
			msgTimestamp := messageSentBuffer[i][:19]
			intTimestamp, _ := strconv.ParseInt(string(msgTimestamp), 10, 64)
			if intTimestamp+TIMEOUT < time.Now().UnixNano() { //il faut renvoyer le paquet, il est timeout
				//fmt.Print("renvoi du paquet : ")
				//fmt.Println(string(messageSentBuffer[i][19:25]))
				_, err := socketCommunication.WriteToUDP(messageSentBuffer[i][19:], clientAddress)
				if err != nil {
					fmt.Println(err)
					return -1
				}
			}

			select {
			case numAckReceived := <-channelAck:
				for i := 0; i < len(messageSentBuffer); i++ {
					//étape 1: on check les numACK dès qu'on trouve celui qui corrrespond, on supprime le message du buffer, il ne sert plus à rien
					extractNumAck := messageSentBuffer[i][19:25]
					intNumAck, _ := strconv.Atoi(string(extractNumAck))
					if intNumAck <= numAckReceived {
						//fmt.Print("Suppression du buffer : ")
						//fmt.Println(intNumAck)
						messageSentBuffer = append(messageSentBuffer[:i], messageSentBuffer[i+1:]...) //on retire le message qui a été acquitté

						packetCount--
					}
				}
			default:
			}
		}

	}
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

	filenameClient := make([]byte, 1024)
	lengthFilenameClient, clientAddress, err := socketCommunication.ReadFromUDP(filenameClient)
	//fmt.Println(string(filenameClient[0:lengthFilenameClient]))

	file, err := os.Open(string(filenameClient[0 : lengthFilenameClient-1]))

	if err != nil {
		fmt.Println(err)
		return
	}

	chanAck := make(chan int)
	chanStop := make(chan bool)

	go send(clientAddress, socketCommunication, file, chanAck, chanStop)
	go receive(chanAck, socketCommunication, chanStop)

}

func main() {
	rand.Seed(time.Now().Unix())

	var wg sync.WaitGroup

	arguments := os.Args
	if len(arguments) < 4 {
		fmt.Println("args : port window RCVSIZE")
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
			//fmt.Println(handshake2)
			_, err = socketConnect.WriteToUDP([]byte(handshake2), clientAddress)

			if err != nil {
				fmt.Println(err)
				return
			}

			//fmt.Println("SYN-ACK")

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
