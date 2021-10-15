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
        perror("Missing args : ./server <port_serv> <UDP_port>");
        return -1;
    }

  struct sockaddr_in adresse, client, adress_udp;
  fd_set sockets; //creation ensemble de descripteurs
  int port= atoi(argv[1]);
  int port_udp = atoi(argv[2]);
  int valid= 1;
  int valid_udp = 1;
  socklen_t alen= sizeof(client);
  char buffer[RCVSIZE];

  //create socket
  int server_desc = socket(AF_INET, SOCK_STREAM, 0);
  int udp_desc = socket(AF_INET, SOCK_DGRAM, 0);

  //handle error
  if (server_desc < 0) {
    perror("Cannot create socket\n");
    return -1;
  }

  setsockopt(server_desc, SOL_SOCKET, SO_REUSEADDR, &valid, sizeof(int));
  setsockopt(udp_desc, SOL_SOCKET, SO_REUSEADDR, &valid_udp, sizeof(int));

  adresse.sin_family= AF_INET;
  adresse.sin_port= htons(port);
  adresse.sin_addr.s_addr= htonl(INADDR_ANY);

  adress_udp.sin_family = AF_INET;
  adress_udp.sin_port= htons(port_udp);
  adress_udp.sin_addr.s_addr = htonl(INADDR_ANY);

  //initialize socket
  if (bind(server_desc, (struct sockaddr*) &adresse, sizeof(adresse)) == -1) {
    perror("Bind failed\n");
    close(server_desc);
    return -1;
  }

  if (bind(udp_desc, (struct sockaddr*) &adress_udp, sizeof(adress_udp)) == -1) {
      perror("Bind UDP failed\n");
      close(udp_desc);
      return -1;
  }


  //listen to incoming clients
  if (listen(server_desc, 0) < 0) {
    printf("Listen failed\n");
    return -1;
  }

  FD_ZERO(&sockets); // on initialise à zéro le set de descripteurs

  printf("Listen done\n");

  while (1) {

    FD_SET(server_desc, &sockets); //on active les bits correspondants aux descripteurs des sockets d'écoute
    FD_SET(udp_desc, &sockets);

    printf("Accepting\n");
    //int client_desc = accept(server_desc, (struct sockaddr*)&client, &alen);
    int socket_activities = select(5, &sockets, NULL, NULL, NULL); //on surveille uniquement l'envoie de flux vers le serveur

    if(FD_ISSET(server_desc, &sockets)==1){

        pid_t pid_val = fork(); //on fork --> création d'un processus fils identique au père
        printf("%d \n", pid_val);

        int client_desc = accept(server_desc, (struct sockaddr*)&client, &alen);

        int msgSize = read(client_desc,buffer,RCVSIZE);

        while (msgSize > 0) {
            write(client_desc,buffer,msgSize);
            printf("%s",buffer);
            memset(buffer,0,RCVSIZE);
            msgSize = read(client_desc,buffer,RCVSIZE);
        }

        close(client_desc);

        if(pid_val == 0){
            printf("Fin du process fils\n");
            exit(0);
        }

    }else if(FD_ISSET(udp_desc, &sockets) == 1){
        int len = sizeof(adress_udp);
        ssize_t received = recvfrom(udp_desc, (char *)buffer, RCVSIZE, MSG_WAITALL, (struct sockaddr *) &adress_udp, &len);
        printf("Client > %s", buffer);
        sendto(udp_desc, (const char *)buffer, RCVSIZE, MSG_CONFIRM, (const struct sockaddr *) &adress_udp, len);
        printf("Message resent to the client");
    }

    /*if(pid_val == 0) {
        printf("Waiting socket processus fils : %d \n", server_desc);
        printf("Accept socket processus fils : %d\n", client_desc);
        close(server_desc);
    }else{
        printf("Waiting socket processus pere : %d \n", server_desc);
        printf("Accept socket processus pere : %d\n", client_desc);
        close(client_desc); //on close la socket acceptation de client dans le process père
    }*/
    


  }

close(server_desc);
return 0;
}
