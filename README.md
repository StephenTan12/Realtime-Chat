# Realtime Chat

## Description

This simulates a Realtime Chat server and client in a local environment. This was done by using TCP sockets, where in the case where the server receives a message from one of the connected clients, the server would broadcast the message to all connected clients.

## Usage

### Run
1) Build files
```
bash start.sh
```
2) Start server
```
./realchat server
```
3) Connect to server 
```
./realchat client
```