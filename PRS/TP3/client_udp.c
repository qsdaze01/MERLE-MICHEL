#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>

#define RCVSIZE 1024

int main (int argc, char *argv[]) {

    if(argc != 4){
        perror("Missing args : ./client <server_address> <server_port> <file name>");
        return -1;
    }

    struct sockaddr_in adresse;
    struct sockaddr_in adresse_com;
    int port = atoi(argv[2]);
    int valid = 1;
    int word = 0;
    int port_communication;
    char *handshake_1 = "SYN";
    char handshake_2[RCVSIZE];
    char *handshake_3 = "ACK";
    char *ptr_word[RCVSIZE];
    char *fileName = argv[3];

    //create socket
    int server_desc = socket(AF_INET, SOCK_DGRAM, 0);

    // handle error
    if (server_desc < 0) {
        perror("cannot create socket\n");
        return -1;
    }

    setsockopt(server_desc, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));

    adresse.sin_family= AF_INET;
    adresse.sin_port= htons(port);
    adresse.sin_addr.s_addr= htonl(atoi(argv[1]));

    socklen_t len = sizeof(adresse);
    
    sendto(server_desc, (const char *)handshake_1, RCVSIZE, 0, (const struct sockaddr *) &adresse, len);

    int taille = recvfrom(server_desc, handshake_2, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse, &len);
    if(taille < 0){
        perror("Erreur receive");
    }

    //Récupération handshake_2
    ptr_word[word] = strtok(handshake_2," ");
    while(ptr_word[word] != NULL){
        printf("%s\n", ptr_word[word]);
        word++;
        ptr_word[word] = strtok(NULL, " ");
    }
    word = 0;

    if(strcmp(ptr_word[0], "SYN") == 0){
        if(strcmp(ptr_word[1], "ACK") == 0){
            port_communication = atoi(ptr_word[word]);
        }else{
            fprintf(stderr, "Erreur handshake \n");
            return(-1);
        }
    }else{
        fprintf(stderr, "Erreur handshake \n");
        return(-1);
    }

    sendto(server_desc, (const char *)handshake_3, RCVSIZE, 0, (const struct sockaddr *) &adresse, len);

    int com_desc = socket(AF_INET, SOCK_DGRAM, 0);
    setsockopt(com_desc, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));

    adresse_com.sin_family = AF_INET;
    adresse_com.sin_port= htons(port_communication);
    adresse_com.sin_addr.s_addr = htonl(INADDR_ANY);
                    
    if (bind(com_desc, (struct sockaddr*) &adresse_com, sizeof(adresse_com)) == -1) {
        perror("Bind UDP failed\n");
        close(com_desc);
        return -1;
    }

    sendto(com_desc, (const char *)fileName, RCVSIZE, 0, (const struct sockaddr *) &adresse_com, len);


    while (1) {
        //sendto(server_desc, (const char *)msg, RCVSIZE, 0, (const struct sockaddr *) &adresse, len);
    }

    return 0;
}

