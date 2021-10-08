#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>

int main(){

    int socketServeur = socket(AF_INET, SOCK_STREAM, 0);
    if(socketServeur == -1){
        printf("Erreur socket \n");
        return(-1);
    }

    int reuse = 1;
    setsockopt(socketServeur, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));
    printf("Valeur socket Client : %d \n", socketServeur);

    struct sockaddr_in addr_serv;
    memset((char*) &addr_serv, 0, sizeof(addr_serv));

    addr_serv.sin_addr.s_addr = INADDR_ANY;
    char addr_any[INET_ADDRSTRLEN];
    printf("Valeur retournée par aton : %d \n", inet_aton("0.0.0.0", &addr_serv.sin_addr));
    inet_ntop(AF_INET, &addr_serv.sin_addr.s_addr, addr_any, sizeof(addr_any));
    printf("Adresse Serveur : %s \n", addr_any);

    addr_serv.sin_port = htons(1234);
    addr_serv.sin_family = AF_INET;

    if(bind(socketServeur, (struct sockaddr *) &addr_serv, sizeof(addr_serv)) == -1){
        printf("bind \n");
        return(-1);
    }

    if(listen(socketServeur, 10) != 0){
        printf("listen \n");
        return(-1);
    }

    struct sockaddr_in addr_client;
    memset((char*) &addr_client, 0, sizeof(addr_client));

    int sizeClient = sizeof(addr_client);
    int socketClient = accept(socketServeur, (struct sockaddr *) &addr_client, (socklen_t *) &sizeClient);
    printf("%d \n", socketClient);

    char message[3];
    message[0] = 'a';
    message[1] = 'z';
    char buffer[3];
    while(1){
        write(socketClient, message, sizeof(message)-1);
        if(read(socketClient, buffer, sizeof(buffer)-1) == -1){
            printf("error \n");
            return(-1);
        }
        printf("%s \n", buffer);

    }
    
    return(0);
}