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

		filename := []byte("test.pdf")
		_, err = socketCommunication.Write(filename)

		if err != nil {
			fmt.Println(err)
			return
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
