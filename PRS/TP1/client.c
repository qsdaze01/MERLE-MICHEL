#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>

int main(){
    
    int socketClient = socket(AF_INET, SOCK_STREAM, 0);
    
    if(socketClient == -1){
        printf("Erreur socket \n");
        return(-1);
    }

    int reuse = 1;
    setsockopt(socketClient, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));

    printf("%d \n", socketClient);

    return(0);
}