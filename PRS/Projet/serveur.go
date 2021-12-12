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

type messageBuffer struct {
	timestamp int64
	message   []byte
	numSeq    int
}

var arg = os.Args
var RCVSIZE = 1500
var RTT = 4
var TIMEOUT = int64(RTT * 1000000)

var window = 50

var sleep = 0

//go routine permettant de recevoir en permanence les ack venant du client et de les envoyer à la go routine send pour qu'elle puisse gérer les retransmissions
func receive(channelAck chan int, socketCommunication *net.UDPConn) {
	for {
		messageAck := make([]byte, 10)
		_, _, _ = socketCommunication.ReadFromUDP(messageAck)

		numAck := string(messageAck[3:9])
		res, _ := strconv.Atoi(numAck)

		channelAck <- res
	}
}

func send(clientAddress *net.UDPAddr, socketCommunication *net.UDPConn, file *os.File, channelAck chan int, chanStop chan int) int {
	seq := []byte("000001")
	numSeq, _ := strconv.Atoi(string(seq))
	packetCount := 0
	startTimer := time.Now()
	var numSeqEndOfFile = 0
	var numAckReceived = -1
	var endOfFile = false
	var numAck = -1
	var numAckCount = 1
	fileBuffer := make([]byte, RCVSIZE-6)
	messageMap := make(map[int]messageBuffer) //création d'un buffer sous forme d'une map
	var numAckDeleted = 1

	for {
		for (packetCount < window) && (endOfFile == false) {

			offset := (int64)((numSeq - 1) * (RCVSIZE - 6))
			bytesRead, err := file.ReadAt(fileBuffer, offset)
			if err == io.EOF {
				endOfFile = true //permet de ne plus rentrer dans la boucle en cas d'eof puisqu'on a plus besoin de lire le fichier
				numSeqEndOfFile = numSeq

				elem := messageBuffer{}
				elem.timestamp = time.Now().UnixNano()
				elem.message = append(seq, fileBuffer...)
				elem.numSeq = numSeq
				messageMap[numSeq] = elem

				_, err := socketCommunication.WriteToUDP(elem.message[:bytesRead+6], clientAddress)
				if err != nil {
					//fmt.Println(err)
					return -1
				}

			} else if err != nil {
				fmt.Println(err)
				return -1
			} else {
				elem := messageBuffer{}
				elem.timestamp = time.Now().UnixNano()
				elem.message = append(seq, fileBuffer...)
				elem.numSeq = numSeq
				messageMap[numSeq] = elem //on place l'élément à la fin pour plus tard limiter le nombre d'itérations sur la boucle for : les plus anciens seront au début de la linkedlist
				_, err := socketCommunication.WriteToUDP(elem.message[:bytesRead+6], clientAddress)
				if err != nil {
					//fmt.Println(err)
					return -1
				}
			}

			packetCount++

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

		if (endOfFile == true) && (numAckReceived == numSeqEndOfFile) { //quand on a reçu l'acquittement du dernier paquet, on peut envoyer FIN
			eof := make([]byte, 3)
			eof[0] = byte('F')
			eof[1] = byte('I')
			eof[2] = byte('N')
			endTimer := time.Now()
			diffTimer := endTimer.Sub(startTimer)
			fmt.Println("EOF envoyé, fichier transféré avec succès !")
			fmt.Println(diffTimer)
			for i := 0; i < 100; i++ {
				_, _ = socketCommunication.WriteToUDP(eof, clientAddress)
			}
			chanStop <- 1 //on dit à la goroutine receive de s'arrêter aussi
			return 0      //on s'arrête quand on a tout reçu
		}

		select {
		case numAckReceived = <-channelAck:
			if numAckReceived == numAck { //pour fast retransmit
				numAckCount++ //on incrémente le compteur des duplicate ack
			} else {
				packets := numAckReceived - numAck
				packetCount -= packets
				if packetCount < 0 { //TODO: Essayer en le retirant si besoin
					packetCount = 0 //pour éviter qu'on dépasse la fenêtre
				}
				if numAckReceived != 0 { //go routine receive renvoit 0 si elle est en timeout
					numAck = numAckReceived //nouvel ack reçu on remet tout à 0
					numAckCount = 1
				}

			}

		default:
		}

		for i := numAckDeleted; i <= numAck; i++ { //permet d'optimiser la recherche aux numéros ack présents effectivement dans la map
			if _, ok := messageMap[i]; ok {
				delete(messageMap, i) //on supprime les messages acquittés
			}
			numAckDeleted++
		}

		for _, value := range messageMap {
			if time.Now().UnixNano()-value.timestamp > TIMEOUT {
				_, err := socketCommunication.WriteToUDP(value.message, clientAddress)
				if err != nil {
					fmt.Println(err)
					return 0
				}

				if sleep != 0 {
					time.Sleep(time.Duration(sleep) * time.Microsecond)
				}
			}

			select {
			case numAckReceived = <-channelAck:
				if numAckReceived == numAck { //pour fast retransmit
					numAckCount++ //on incrémente le compteur des duplicate ack
				} else {
					packets := numAckReceived - numAck
					packetCount -= packets
					if packetCount < 0 { //TODO: Essayer en le retirant si besoin
						packetCount = 0 //pour éviter qu'on dépasse la fenêtre
					}
					if numAckReceived != 0 { //go routine receive renvoit 0 si elle est en timeout
						numAck = numAckReceived //nouvel ack reçu on remet tout à 0
						numAckCount = 1
					}

				}

			default:
			}
		}

		for i := numAckDeleted; i <= numAck; i++ { //permet d'optimiser la recherche aux numéros ack présents effectivement dans la map
			if _, ok := messageMap[i]; ok {
				delete(messageMap, i) //on supprime les messages acquittés
			}
			numAckDeleted++
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

	file, err := os.Open(string(filenameClient[0 : lengthFilenameClient-1]))

	if err != nil {
		fmt.Println(err)
		return
	}

	chanAck := make(chan int)
	chanStop := make(chan int)

	go send(clientAddress, socketCommunication, file, chanAck, chanStop)
	go receive(chanAck, socketCommunication)
}

func main() {
	rand.Seed(time.Now().Unix())

	var wg sync.WaitGroup

	arguments := os.Args
	if len(arguments) < 2 {
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
			_, err = socketConnect.WriteToUDP([]byte(handshake2), clientAddress)

			if err != nil {
				fmt.Println(err)
				return
			}

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
