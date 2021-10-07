all: client serveur

client: client.o 
	gcc client.o -Wall -o client

client.o: client.c sys/types.h sys/socket.h
	gcc -c client.c -Wall -o client.o

serveur: serveur.o
	gcc serveur.o -Wall -o serveur

serveur.o: serveur.c sys/types.h sys/socket.h
	gcc -c serveur.c -Wall -o serveur.o

clean:
	rm client *.o serveur