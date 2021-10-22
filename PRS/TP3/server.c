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

    if(argc != 3){
        perror("Missing args : ./server <UDP_port> <port_communication>");
        return -1;
    }

    struct sockaddr_in adresse_udp;
    struct sockaddr_in adresse_com;
    fd_set sockets; //creation ensemble de descripteurs
    int port_udp = atoi(argv[1]);
    int port_communication = atoi(argv[2]);
    int valid_udp = 1;
    char handshake_1[RCVSIZE];
    char handshake_2[RCVSIZE] = "SYN ACK ";
    char handshake_3[RCVSIZE];
    strcat(handshake_2, argv[2]);
    char fileName[RCVSIZE];

    //create socket
    int udp_desc = socket(AF_INET, SOCK_DGRAM, 0);

    setsockopt(udp_desc, SOL_SOCKET, SO_REUSEADDR, &valid_udp, sizeof(int));

    adresse_udp.sin_family = AF_INET;
    adresse_udp.sin_port= htons(port_udp);
    adresse_udp.sin_addr.s_addr = htonl(INADDR_ANY);

    //initialize socket
    if (bind(udp_desc, (struct sockaddr*) &adresse_udp, sizeof(adresse_udp)) == -1) {
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
        select(5, &sockets, NULL, NULL, NULL); //on surveille uniquement l'envoie de flux vers le serveur

        if(FD_ISSET(udp_desc, &sockets) == 1){
            socklen_t len = sizeof(adresse_udp);
            recvfrom(udp_desc, (char *)handshake_1, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_udp, &len);                        
            if(strcmp(handshake_1, "SYN") == 0){
                printf("OK 1 \n");
                
                int com_desc = socket(AF_INET, SOCK_DGRAM, 0);
                setsockopt(com_desc, SOL_SOCKET, SO_REUSEADDR, &valid_udp, sizeof(int));

                adresse_com.sin_family = AF_INET;
                adresse_com.sin_port= htons(port_communication);
                adresse_com.sin_addr.s_addr = htonl(INADDR_ANY);
                    
                if (bind(com_desc, (struct sockaddr*) &adresse_com, sizeof(adresse_com)) == -1) {
                    perror("Bind UDP failed\n");
                    close(com_desc);
                    return -1;
                }
                //mettre le fork()

                sendto(udp_desc, (char *)handshake_2, RCVSIZE, MSG_CONFIRM, (struct sockaddr *) &adresse_udp, len);               
                recvfrom(udp_desc, (char *)handshake_3, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_udp, &len);            
                if(strcmp(handshake_3, "ACK") == 0){
                    printf("OK \n");
                    recvfrom(com_desc, (char *)fileName, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_com, &len);            
                }
            }

            printf("Message resent to the client \n");
        }

    }

    close(udp_desc);
    return 0;
}
