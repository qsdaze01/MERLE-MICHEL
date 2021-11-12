package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	RCVSIZE := 10000
	arguments := os.Args
	if len(arguments) == 2 {
		fmt.Println("args : address port")
		return
	}

	connectionParameters := arguments[1] + ":" + arguments[2]

	addressPort, err := net.ResolveUDPAddr("udp4", connectionParameters)
	socketConnect, err := net.DialUDP("udp4", nil, addressPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The UDP server is %s\n", socketConnect.RemoteAddr().String())

	handshake1 := []byte("SYN")
	_, err = socketConnect.Write(handshake1)

	if err != nil {
		fmt.Println(err)
		return
	}

	handshake2 := make([]byte, 1024)
	lengthHandshake2, _, err := socketConnect.ReadFromUDP(handshake2)

	if err != nil {
		fmt.Println(err)
		return
	}

	splitHandshake2 := strings.Split(string(handshake2[0:lengthHandshake2]), " ")

	if splitHandshake2[0] == "SYN_ACK" {
		communicationParameters := arguments[1] + ":" + splitHandshake2[1]
		fmt.Println(communicationParameters)
		communicationAddressPort, err := net.ResolveUDPAddr("udp4", communicationParameters)
		socketCommunication, err := net.DialUDP("udp4", nil, communicationAddressPort)

		if err != nil {
			fmt.Println(err)
			return
		}

		defer socketCommunication.Close()

		handshake3 := []byte("ACK")
		_, err = socketConnect.Write(handshake3)

		if err != nil {
			fmt.Println(err)
			return
		}

		err = socketConnect.Close()

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Handshaked !")

		filename := []byte("hawkeye.jpg")
		_, err = socketCommunication.Write(filename)

		if err != nil {
			fmt.Println(err)
			return
		}

		message := make([]byte, RCVSIZE)
		fileBuffer := make([]byte, RCVSIZE-10)
		file, err := os.OpenFile("received.jpg", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)

		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			_, _, err := socketCommunication.ReadFromUDP(message)
			for i := range fileBuffer {
				fileBuffer[i] = message[i]
			}

			fmt.Println("Message :" + string(message))

			if err != nil {
				fmt.Println(err)
				return
			}
			ack := []byte("ACK ")
			tabSeq := make([]byte, 10)
			for i := 0; i < 10; i++ {
				tabSeq[i] = message[i+len(fileBuffer)]
			}

			messageAck := append(ack, tabSeq...)
			//fmt.Println("ACK :" + string(messageAck))
			_, err = socketCommunication.Write(messageAck)

			if err != nil {
				fmt.Println(err)
				return
			}

			//fmt.Println("Buffer :" + string(fileBuffer))
			if string(fileBuffer[0:3]) == "EOF" {
				fmt.Println("bouh")
				break //End of File
			}

			_, err = file.Write(fileBuffer) //TODO: pb ici quand une partie des cases du buffer sont vides (fin de fichier) il write des caractères NUL dans le fichier il faut trouver une condition

			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
