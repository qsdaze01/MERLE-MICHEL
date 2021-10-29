#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <sys/types.h>
#include <sys/select.h>
#include <sys/time.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

#define RCVSIZE 1024

int main (int argc, char *argv[]) {

    if(argc != 3){
        perror("Missing args : ./server <UDP_port> <port_communication>");
        return -1;
    }

    /* Déclaration des variables */
    struct sockaddr_in adresse_udp;
    struct sockaddr_in adresse_com;
    fd_set sockets; //creation ensemble de descripteurs
    int port_udp = atoi(argv[1]);
    int port_communication = atoi(argv[2]);
    int valid_udp = 1;
    int word = 0;
    double RTT = 0;
    char handshake_1[RCVSIZE];
    char handshake_2[RCVSIZE] = "SYN ACK ";
    char handshake_3[RCVSIZE];
    char *ptr_word[RCVSIZE];
    strcat(handshake_2, argv[2]);
    char fileName[RCVSIZE];
    clock_t begin = clock();
    clock_t end = clock();
    srand(time(NULL));

    /* Création de la socket de connexion au serveur */
    int udp_desc = socket(AF_INET, SOCK_DGRAM, 0);

    setsockopt(udp_desc, SOL_SOCKET, SO_REUSEADDR, &valid_udp, sizeof(int));

    adresse_udp.sin_family = AF_INET;
    adresse_udp.sin_port= htons(port_udp);
    adresse_udp.sin_addr.s_addr = htonl(INADDR_ANY);

    /* Initialisation de la socket */
    if (bind(udp_desc, (struct sockaddr*) &adresse_udp, sizeof(adresse_udp)) == -1) {
        perror("Bind UDP failed\n");
        close(udp_desc);
        return -1;
    }

    FD_ZERO(&sockets); // on initialise à zéro le set de descripteurs

    while (1) {

        /* Sélection de la socket */
        //on active les bits correspondants aux descripteurs des sockets d'écoute
        FD_SET(udp_desc, &sockets);

        printf("Accepting\n");
        select(5, &sockets, NULL, NULL, NULL); //on surveille uniquement l'envoie de flux vers le serveur

        if(FD_ISSET(udp_desc, &sockets) == 1){
            socklen_t len = sizeof(adresse_udp);

            /* Handshake */
            recvfrom(udp_desc, (char *)handshake_1, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_udp, &len);                        
            if(strcmp(handshake_1, "SYN") == 0){
                handshake_1[0] = '\0';
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
                begin = clock();
                sendto(udp_desc, (char *)handshake_2, RCVSIZE, MSG_CONFIRM, (struct sockaddr *) &adresse_udp, len);               
                recvfrom(udp_desc, (char *)handshake_3, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_udp, &len);
                end = clock();
                RTT = 4*((double)(end-begin)/CLOCKS_PER_SEC);
                printf("RTT : %f \n", RTT);

                if(strcmp(handshake_3, "ACK") == 0){
                    handshake_3[0] = '\0';
                    pid_t pid_val = fork();
                    printf("%d \n", pid_val);
                    if(pid_val == 0){

                        /* Process fils - communication avec le client */
                        close(udp_desc);
                        printf("%d %s %d %d \n", com_desc, inet_ntoa(adresse_com.sin_addr), ntohs(adresse_com.sin_port),len);
                        recvfrom(com_desc, (char *)fileName, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adresse_com, &len);
                        printf("%s\n", fileName);

                        /* Ajout timeout */
                        struct timeval tv;
                        tv.tv_sec = 0;
                        tv.tv_usec = (int) (100000*RTT);
                        if (setsockopt(com_desc, SOL_SOCKET, SO_RCVTIMEO,&tv,sizeof(tv)) < 0) {
                            perror("Error");
                        }

                        /* PARTIE LECTURE DE FICHER */
                        //Problème : il relit certains bit déjà lus et fin des bits un peu chelou
                        FILE* fichierClient = fopen(fileName, "rb");
                            if(fichierClient == NULL){
                            printf("Erreur fichier \n");
                            return(-1);
                        }

                        char buffer_fichier[RCVSIZE];
                        int num_seq = 0.1*rand();

                        while(feof(fichierClient) == 0){
                            fread((void *) buffer_fichier, 1, RCVSIZE - 10, fichierClient);

                            char seq[10];
                            sprintf(seq, "%d", num_seq);
                            printf("%s \n", seq);
                            for(int i = 0; i < 10; i++){
                                buffer_fichier[RCVSIZE - 10 + i] = seq[i];
                            }

                            printf("%s\n", buffer_fichier);


                            int taille_envoi_fichier = sendto(com_desc, (char *)buffer_fichier, RCVSIZE, MSG_CONFIRM, (struct sockaddr *) &adresse_com, len);
                            //printf("Ok \n");

                            if(taille_envoi_fichier < 0){
                                perror("Erreur envoi fichier");
                            }
                            char ack[RCVSIZE];

                            if(recvfrom(com_desc, (char *) ack , RCVSIZE, 0, (struct sockaddr *) &adresse_com, &len) < 0){
                                //timeout reached
                                printf("Timout reached. Resending segment\n");
                            }
                            //taille_reception = recvfrom(com_desc, (char *)ack, RCVSIZE, MSG_DONTWAIT, (struct sockaddr *) &adresse_com, &len);

                            ptr_word[word] = strtok(ack," ");
                            while(ptr_word[word] != NULL){
                                word++;
                                ptr_word[word] = strtok(NULL, " ");
                            }
                            word = 0;

                            if(strcmp(ptr_word[0], "ACK") != 0){
                                printf("pas de ack reçu \n");
                                return(-1);
                            }
                            //RTT = 4*difftime(end, begin);

                            printf("ack reçu num_seq = %s \n", ptr_word[1]);
                            num_seq++;
                        }

                        //On tue le process enfant
                        close(com_desc);
                        return(0);
                    }else{

                        /* Process père - on reboucle pour écouter un autre client */
                        close(com_desc);
                        port_communication++;
                    }
                }
            }
        }
    }

    close(udp_desc);
    return 0;
}
