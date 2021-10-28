#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define RCVSIZE 1024

int main (int argc, char *argv[]) {

    if(argc != 4){
        perror("Missing args : ./client <server_address> <server_port> <file name>");
        return -1;
    }

    /* Déclaration des variables */
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

    /* Création de la socket de connection au serveur */
    int server_desc = socket(AF_INET, SOCK_DGRAM, 0);


    if (server_desc < 0) {
        perror("cannot create socket\n");
        return -1;
    }

    setsockopt(server_desc, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));

    adresse.sin_family= AF_INET;
    adresse.sin_port= htons(port);
    inet_aton(argv[1], &adresse.sin_addr);

    socklen_t len = sizeof(adresse);

    /* Handshake */
    sendto(server_desc, (const char *)handshake_1, RCVSIZE, 0, (const struct sockaddr *) &adresse, len);

    int taille = recvfrom(server_desc, handshake_2, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse, &len);
    if(taille < 0){
        perror("Erreur receive");
    }

    ptr_word[word] = strtok(handshake_2," ");
    while(ptr_word[word] != NULL){
        word++;
        ptr_word[word] = strtok(NULL, " ");
    }
    word = 0;

    if(strcmp(ptr_word[0], "SYN") == 0){
        if(strcmp(ptr_word[1], "ACK") == 0){
            port_communication = atoi(ptr_word[2]);
        }else{
            fprintf(stderr, "Erreur handshake \n");
            return(-1);
        }
    }else{
        fprintf(stderr, "Erreur handshake \n");
        return(-1);
    }

    sendto(server_desc, (const char *)handshake_3, RCVSIZE, 0, (const struct sockaddr *) &adresse, len);
    /* Fin du handshake */

    /* Création de la socket de communcation */
    int com_desc = socket(AF_INET, SOCK_DGRAM, 0);
    setsockopt(com_desc, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));

    adresse_com.sin_family = AF_INET;
    adresse_com.sin_port= htons(port_communication);
    inet_aton(argv[1], &adresse_com.sin_addr);
                    
    //printf("%d %s %d %d \n", com_desc, inet_ntoa(adresse_com.sin_addr), ntohs(adresse_com.sin_port),len);

    /* Envoi du nom du fichier */
    int taille_fileName = sendto(com_desc, (const char *)fileName, RCVSIZE, 0, (const struct sockaddr *) &adresse_com, len);

    if(taille_fileName < 0){
        perror("Erreur envoi");
    }

    /* Réception du fichier */
    char buffer_reception_fichier[RCVSIZE];
    int taille_reception_fichier = recvfrom(com_desc, buffer_reception_fichier, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_com, &len);
    char num_seq[10];
    char ack[RCVSIZE] = "ACK ";
    for(int i = 0; i < 10; i++){
        num_seq[i] = buffer_reception_fichier[RCVSIZE - 10 + i];
    }

    printf("%s \n", num_seq);

    strcat(ack, num_seq);
    sendto(com_desc, (const char *)ack, RCVSIZE, 0, (const struct sockaddr *) &adresse_com, len);
    while(taille_reception_fichier > 0){
        *ack = "ACK";
        printf("préOK \n");
        taille_reception_fichier = recvfrom(com_desc, buffer_reception_fichier, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_com, &len);
        printf("Ok \n");

        for(int i = 0; i < 10; i++){
            num_seq[i] = buffer_reception_fichier[RCVSIZE - 10 + i];
        }
        strcat(ack, num_seq);
        sendto(com_desc, (const char *)ack, RCVSIZE, 0, (const struct sockaddr *) &adresse_com, len);
    }

    return 0;
}

