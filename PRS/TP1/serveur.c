#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>

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
    printf("Valeur retournée par aton : %d \n", inet_aton("0.0.0.0", (struct in_addr *) &addr_serv.sin_addr.s_addr));
    inet_ntop(AF_INET, &addr_serv.sin_addr.s_addr, addr_any, sizeof(addr_any));
    printf("Adresse Serveur : %s \n", addr_any);

    addr_serv.sin_port = 1234;

    if(bind(socketServeur, (struct sockaddr *) &addr_serv.sin_addr.s_addr, (socklen_t) sizeof(addr_serv.sin_addr.s_addr)) == -1){
        printf("bind \n");
        return(-1);
    }

    if(listen(socketServeur, 10) == -1){
        printf("listen \n");
        return(-1);
    }

    while(1){
        int a = accept(socketServeur, (struct sockaddr *) 0, (socklen_t *) 0);
        printf("%d \n", a);
        /*if(accept(socketServeur, (struct sockaddr *) 0, (socklen_t *) 0) != -1){
            printf("a \n");
        }*/
    }

    return(0);
}