#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/select.h>
#include <sys/time.h>
#include <sys/socket.h>
#include <netinet/in.h>

#define RCVSIZE 1024

int main (int argc, char *argv[]) {

    if(argc != 2){
        perror("Missing args : ./server <UDP_port>");
        return -1;
    }

    struct sockaddr_in adresse, adress_udp;
    fd_set sockets; //creation ensemble de descripteurs
    int port_udp = atoi(argv[1]);
    int valid= 1;
    int valid_udp = 1;
    socklen_t alen= sizeof(client);
    char buffer[RCVSIZE];

    //create socket
    int udp_desc = socket(AF_INET, SOCK_DGRAM, 0);

    setsockopt(udp_desc, SOL_SOCKET, SO_REUSEADDR, &valid_udp, sizeof(int));

    adress_udp.sin_family = AF_INET;
    adress_udp.sin_port= htons(port_udp);
    adress_udp.sin_addr.s_addr = htonl(INADDR_ANY);

    //initialize socket
    if (bind(udp_desc, (struct sockaddr*) &adress_udp, sizeof(adress_udp)) == -1) {
        perror("Bind UDP failed\n");
        close(udp_desc);
        return -1;
    }
    FD_ZERO(&sockets); // on initialise à zéro le set de descripteurs

    printf("Listen done\n");

    while (1) {

        //on active les bits correspondants aux descripteurs des sockets d'écoute
        FD_SET(udp_desc, &sockets);

        printf("Accepting\n");
        //int client_desc = accept(server_desc, (struct sockaddr*)&client, &alen);
        int socket_activities = select(5, &sockets, NULL, NULL, NULL); //on surveille uniquement l'envoie de flux vers le serveur

        if(FD_ISSET(udp_desc, &sockets) == 1){
            int len = sizeof(adress_udp);
            ssize_t received = recvfrom(udp_desc, (char *)buffer, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adress_udp, &len);
            printf("Client > %s", buffer);
            sendto(udp_desc, (const char *)buffer, RCVSIZE, MSG_CONFIRM, (const struct sockaddr *) &adress_udp, len);
            printf("Message resent to the client");
        }

    }

    close(udp_desc);
    return 0;
}
