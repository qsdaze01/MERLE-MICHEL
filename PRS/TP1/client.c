#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>

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

    //addr_client.sin_addr.s_addr = INADDR_LOOPBACK;
    char addr_any[INET_ADDRSTRLEN];
    printf("Valeur retournée par aton : %d \n", inet_aton("127.0.0.1", &addr_client.sin_addr));
    inet_ntop(AF_INET, &addr_client.sin_addr.s_addr, addr_any, sizeof(addr_any));
    printf("Adresse Client : %s \n", addr_any);

    addr_client.sin_port = htons(1234);
    addr_client.sin_family = AF_INET;

    connect(socketClient, (struct sockaddr *) &addr_client, sizeof(addr_client));

    char buffer[3];
    char message[3];
    message[0] = 'e';
    message[1] = 'r';
    while(1){
        if(read(socketClient, buffer, sizeof(buffer)-1) == -1){
            printf("error");
            return(-1);
        }
        printf("%s \n", buffer);
        write(socketClient, message, sizeof(message)-1);
    }
    
    
    return(0);
}