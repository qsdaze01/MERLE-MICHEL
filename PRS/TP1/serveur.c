#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>

int main(){

    int socketServeur = socket(AF_INET, SOCK_STREAM, 0);
    if(socketServeur == -1){
        printf("Erreur socket \n");
        return(-1);
    }

    int reuse = 1;
    setsockopt(socketClient, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));
    printf("%d \n", socketServeur);

    return(0);
}