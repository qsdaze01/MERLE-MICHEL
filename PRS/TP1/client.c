#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>

int main(){
    
    int socketClient = socket(AF_INET, SOCK_STREAM, 0);
    
    if(socketClient == -1){
        printf("Erreur socket \n");
        return(-1);
    }

    int reuse = 1;
    setsockopt(socketClient, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));
    printf("Valeur socket Client : %d \n", socketClient);

    struct sockaddr_in addr_client;
    memset((char*) &addr_client, 0, sizeof(addr_client));

    addr_client.sin_addr.s_addr = INADDR_ANY;
    char addr_any[INET_ADDRSTRLEN];
    printf("Valeur retournée par aton : %d \n", inet_aton("127.0.0.2", &addr_client.sin_addr.s_addr));
    inet_ntop(AF_INET, &addr_client.sin_addr.s_addr, addr_any, sizeof(addr_any));
    printf("Adresse Client : %s \n", addr_any);

    addr_client.sin_port = 1234;

    struct sockaddr_in addr_serv;
    memset((char*) &addr_serv, 0, sizeof(addr_serv));   

    addr_serv.sin_addr.s_addr = INADDR_ANY;
    connect(socketClient, &addr_serv.sin_addr.s_addr, sizeof(addr_serv.sin_addr.s_addr));

    return(0);
}