# MERLE-MICHEL    

## ToDo list :    
- Problème EOF côté serveur qui continue de renvoyer le dernier paquet vu qu'il ne reçoit pas d'ACK de la part du client
- Problème des mutex quand on a plusieurs client (on essaye de release un mutex déjà unlocked)
- Mettre un lock et faire un tableau pour les window
- Faire des tests avec un gros fichier



## Astuce FLM

- On aura peut-être pas à fragmenter de notre côté, on envoie un gros paquet en TCP et on laisse Ethernet fragmenter de son côté

