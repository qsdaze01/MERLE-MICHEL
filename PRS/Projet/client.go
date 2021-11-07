package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
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

		filename := []byte("test.txt")
		_, err = socketCommunication.Write(filename)

		if err != nil {
			fmt.Println(err)
			return
		}

		fileBuffer := make([]byte, 32)
		file, err := os.OpenFile("received.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)

		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			lengthFileBuffer, _, err := socketCommunication.ReadFromUDP(fileBuffer)

			if err != nil {
				fmt.Println(err)
				return
			}

			if string(fileBuffer[0:lengthFileBuffer]) == "EOF" {
				break //End of File
			}

			_, err = file.Write(fileBuffer) //TODO: pb ici quand une partie des cases du buffer sont vides (fin de fichier) il write des caractères NUL dans le fichier il faut trouver une condition

			if err != nil {
				fmt.Println(err)
				return
			}
		}

		/*buffer := make([]byte, 1024)
		n, _, err := socketCommunication.ReadFromUDP(buffer)
		fmt.Println("where is wsh ?")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(buffer[0:n]))*/

	}

}
