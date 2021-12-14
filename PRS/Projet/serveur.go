package main

import (
	"fmt"
	"io"
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
var RTT = 0
var TIMEOUT = int64(RTT * 1000000)

var window = 15

var sleep = 250

//go routine permettant de recevoir en permanence les ack venant du client et de les envoyer à la go routine send pour qu'elle puisse gérer les retransmissions
func receive(channelAck chan int, socketCommunication *net.UDPConn) {
	for {
		messageAck := make([]byte, 10)
		_, _, _ = socketCommunication.ReadFromUDP(messageAck)

		numAck := string(messageAck[3:9])
		res, _ := strconv.Atoi(numAck)

		//on push dans le channel le numéro d'ACK reçu en vu d'un traitement par la go routine send
		channelAck <- res
	}
}

//Gère l'envoi et la retransmission des segments vers le client
func send(clientAddress *net.UDPAddr, socketCommunication *net.UDPConn, file *os.File, channelAck chan int) int {
	seq := []byte("000001")
	numSeq, _ := strconv.Atoi(string(seq))
	packetCount := 0
	startTimer := time.Now()
	var numSeqEndOfFile = 0
	var numAckReceived = -1
	var endOfFile = false
	var numAck = -1
	fileBuffer := make([]byte, RCVSIZE-6)
	messageMap := make(map[int]messageBuffer) //création d'un buffer sous forme d'une map
	var numAckDeleted = 1

	for {
		//tant qu'on a pas atteint la fenêtre fixée
		for (packetCount < window) && (endOfFile == false) {
			//lecture sur le disque du segment directement dans le fichier afin de limiter l'occupation mémoire
			//utilisation d'un offset pour positionner le curseur de lecture
			offset := (int64)((numSeq - 1) * (RCVSIZE - 6))
			bytesRead, err := file.ReadAt(fileBuffer, offset)
			if err == io.EOF {
				endOfFile = true //permet de ne plus rentrer dans la boucle en cas d'eof puisqu'on a plus besoin de lire le fichier
				numSeqEndOfFile = numSeq

				elem := messageBuffer{}
				elem.timestamp = time.Now().UnixNano()

				//concaténation du numéro de séquence en format bytes avec le contenu à envoyer
				elem.message = append(seq, fileBuffer...)
				elem.numSeq = numSeq
				messageMap[numSeq] = elem

				//Envoi du fichier via notre socket udp
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

				//Concaténation du numéro de séquence en format bytes avec le contenu à envoyer
				elem.message = append(seq, fileBuffer...)
				elem.numSeq = numSeq

				//on place le segment envoyé dans notre buffer map
				messageMap[numSeq] = elem

				//on envoit le segment en faisant attention à sa taille (nb de bytes lu dans le fichiers + les 6 bytes de séquence)
				_, err := socketCommunication.WriteToUDP(elem.message[:bytesRead+6], clientAddress)
				if err != nil {
					//fmt.Println(err)
					return -1
				}
			}

			//on augmente notre compteur de fenêtre
			packetCount++

			numSeq++

			//On convertit notre numéro de séquence int en 6 bytes pour le prochain tour du boucle
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

		//quand on a reçu l'acquittement du dernier paquet, on peut envoyer FIN
		if (endOfFile == true) && (numAckReceived == numSeqEndOfFile) {
			eof := make([]byte, 3)
			eof[0] = byte('F')
			eof[1] = byte('I')
			eof[2] = byte('N')
			endTimer := time.Now()
			diffTimer := endTimer.Sub(startTimer)
			fmt.Println("EOF envoyé, fichier transféré avec succès !")
			fmt.Println(diffTimer)

			//on envoit 1000x FIN avec un petit délai entre chaque pour être sur que le client reçoive bien la fin de fichier malgré la saturation
			for i := 0; i < 1000; i++ {
				_, _ = socketCommunication.WriteToUDP(eof, clientAddress)
				time.Sleep(100 * time.Microsecond)
			}
			return 0 //on stoppe la go routine quand la transmission est terminée
		}

		select {
		//on récupère l'ack reçu par receive
		case numAckReceived = <-channelAck:
			//on retire les paquets acquittés du compteur pour libérer la fenêtre
			packets := numAckReceived - numAck
			packetCount -= packets
			if packetCount < 0 {
				packetCount = 0 //pour éviter qu'on dépasse la fenêtre
			}

			numAck = numAckReceived

		default:
		}

		for i := numAckDeleted; i <= numAck; i++ { //permet d'optimiser la recherche aux numéros ack présents effectivement dans la map
			if _, ok := messageMap[i]; ok {
				delete(messageMap, i) //on supprime les messages acquittés
			}
			numAckDeleted++
		}

		//retransmission des segments ayant un timestamp trop ieux par rapport à la valeur du timeout fixée
		for _, value := range messageMap {
			if time.Now().UnixNano()-value.timestamp > TIMEOUT {
				_, err := socketCommunication.WriteToUDP(value.message, clientAddress)
				if err != nil {
					fmt.Println(err)
					return 0
				}

				value.timestamp = time.Now().UnixNano() //on remet le timestamp actuel //TODO: voir si c'est mieux commenté ou pas
				time.Sleep(time.Duration(sleep) * time.Microsecond)

			}

			select {
			//on récupère l'ack reçu par receive
			case numAckReceived = <-channelAck:
				//on retire les paquets acquittés du compteur pour libérer la fenêtre
				packets := numAckReceived - numAck
				packetCount -= packets
				if packetCount < 0 {
					packetCount = 0 //pour éviter qu'on dépasse la fenêtre
				}

				numAck = numAckReceived

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
	//la suppression du buffer et le pop des ack depuis la channel sont faits plusieurs fois pour optimiser au mieux les performances

}

//Go routine qui permet de gérer la socket de communication et qui crée les go routines send et receive pour la transmission du fichier
func communicate(wg *sync.WaitGroup, port string) {
	//on libère le worker dès que la go routine se termie ou raise une erreur
	defer wg.Done()

	portCommunication := ":" + port
	//paramètres de la socket de communication
	communicationParameters, err := net.ResolveUDPAddr("udp4", portCommunication)

	if err != nil {
		fmt.Println(err)
		return
	}

	//création de la socket UDP de communication pour transmettre les fichiers
	socketCommunication, err := net.ListenUDP("udp4", communicationParameters)
	if err != nil {
		fmt.Println(err)
		return
	}

	//réception du nom du fichier demandé par le client
	filenameClient := make([]byte, 1024)
	lengthFilenameClient, clientAddress, err := socketCommunication.ReadFromUDP(filenameClient)

	//ouverture du fichier dans le répertoire courant par défaut
	file, err := os.Open(string(filenameClient[0 : lengthFilenameClient-1]))

	if err != nil {
		fmt.Println(err)
		return
	}

	//création de la channel permettant la communication entre les go routines send et receive
	chanAck := make(chan int)

	//création des go routines send et receive pour la transmission du fichier
	go send(clientAddress, socketCommunication, file, chanAck)
	go receive(chanAck, socketCommunication)
}

// Go routine main principale qui gère la socket de connexion initiale
func main() {

	//wait group pour gérer l'attente et l'arrêt de la go routine communicate (synchronisation entre go routines)
	var wg sync.WaitGroup

	arguments := os.Args
	if len(arguments) < 2 {
		fmt.Println("args : port")
		return
	}
	portConnection := ":" + arguments[1]

	//paramètres de connexion de la socket de connexion
	connectionParameters, err := net.ResolveUDPAddr("udp4", portConnection)
	if err != nil {
		fmt.Println(err)
		return
	}

	//création de la socket UDP de connexion
	socketConnect, err := net.ListenUDP("udp4", connectionParameters)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fermeture anticipée de la socket de connexion une fois qu'on ne l'utilisera plus
	defer socketConnect.Close()

	portCommunication := 2000
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		handshake1 := make([]byte, 1024)
		//réception du SYN de la part du client
		_, clientAddress, err := socketConnect.ReadFromUDP(handshake1)
		if string(handshake1[0:3]) == "SYN" {
			//envoi du SYN-ACK suivi du numéro de port de la socket de communication qui servira pour la transmission du fichier
			handshake2 := "SYN-ACK" + strconv.Itoa(portCommunication)
			_, err = socketConnect.WriteToUDP([]byte(handshake2), clientAddress)

			if err != nil {
				fmt.Println(err)
				return
			}

			//ajout d'un worker pour que la goroutine main attende les autres go routine et ne stoppe pas la transmission
			wg.Add(1)
			//création de la go routine communication contenant la socket de communication
			go communicate(&wg, strconv.Itoa(portCommunication))

			handshake3 := make([]byte, 1024)
			//réception du ACK de la part du client
			lengthHandshake3, _, err := socketConnect.ReadFromUDP(handshake3)

			if err != nil {
				fmt.Println(err)
				return
			}

			if string(handshake3[0:lengthHandshake3-1]) == "ACK" {
				fmt.Println("Handshaked !")
			}

			//on incrémente le numéro du port de communication qui sera communiqué aux clients suivants
			portCommunication++
		}
	}
}
