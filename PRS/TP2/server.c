#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/select.h>
#include <sys/time.h>
#include <netinet/in.h>

#define RCVSIZE 1024

int main (int argc, char *argv[]) {

    if(argc != 3){
      printf("Pas le bon nombre de paramètres \n");
      return(-1);
    }

    struct sockaddr_in adresse, client, adresse_UDP;
    fd_set sockets_tab;
    int port = atoi(argv[1]);
    int port_UDP = atoi(argv[2]);
    int valid = 1;
    socklen_t alen = sizeof(client);
    char buffer[RCVSIZE];

    FD_ZERO(&sockets_tab);

    //create socket
    int server_desc = socket(AF_INET, SOCK_STREAM, 0);
    int server_desc_UDP = socket(AF_INET, SOCK_DGRAM, 0);

    //handle error
    if (server_desc < 0) {
      perror("Cannot create socket\n");
      return -1;
    }

    setsockopt(server_desc, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));
    setsockopt(server_desc_UDP, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));

    adresse.sin_family = AF_INET;
    adresse.sin_port = htons(port);
    adresse.sin_addr.s_addr = htonl(INADDR_ANY);

    adresse_UDP.sin_family = AF_INET;
    adresse_UDP.sin_port = htons(port_UDP);
    adresse_UDP.sin_addr.s_addr = htonl(INADDR_ANY);

    //initialize socket
    if (bind(server_desc, (struct sockaddr*) &adresse, sizeof(adresse)) == -1) {
      perror("Bind failed\n");
      close(server_desc);
      return -1;
    }

    if (bind(server_desc_UDP, (struct sockaddr*) &adresse_UDP, sizeof(adresse_UDP)) == -1) {
      perror("Bind failed\n");
      close(server_desc_UDP);
      return -1;
    }


    //listen to incoming clients
    if (listen(server_desc, 0) < 0) {
      printf("Listen failed\n");
      return -1;
    }

    printf("Listen done\n");

    while (1) {

      FD_SET(server_desc_UDP, &sockets_tab);
      FD_SET(server_desc, &sockets_tab);

      printf("Accepting\n");
      select(5, &sockets_tab, NULL, NULL, NULL);

      if(FD_ISSET(server_desc, &sockets_tab) == 1){
        int client_desc = accept(server_desc, (struct sockaddr*)&client, &alen);
        pid_t pid = fork();
        printf("PID : %d \n", pid);

        if(pid == 0){
          printf("socket fils : %d (serveur) %d (client) \n", server_desc, client_desc);
          printf("Value of accept is:%d\n", client_desc);
          close(server_desc);
        }else{
          close(client_desc);
        }
        
        int msgSize = read(client_desc,buffer,RCVSIZE);

        while (msgSize > 0) {
          write(client_desc,buffer,msgSize);
          printf("%s",buffer);
          memset(buffer,0,RCVSIZE);
          msgSize = read(client_desc,buffer,RCVSIZE);
        }

        close(client_desc);
        if(pid == 0){
          printf("Déconnexion client \n");
          exit(0);
        }
      }else if(FD_ISSET(server_desc_UDP, &sockets_tab) == 1){
        
      }  
    }

  close(server_desc);
  return(0);
}
