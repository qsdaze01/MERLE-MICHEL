#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>

#define RCVSIZE 1024

int main (int argc, char *argv[]) {

    if(argc != 3){
        perror("Missing args : ./client <server_address> <server_port>");
        return -1;
    }

    struct sockaddr_in adresse;
    int port = atoi(argv[2]);
    int valid = 1;
    char msg[RCVSIZE];
    char blanmsg[RCVSIZE];

    msg[0]='a';

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


    int cont= 1;
    while (cont) {
        int len = sizeof(adresse);
        sendto(server_desc, (const char *)msg, RCVSIZE, 0, (const struct sockaddr *) &adresse, len);
    }

    return 0;
}

