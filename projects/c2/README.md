# Emulated Sysjoker C2

Reference: https://research.checkpoint.com/2023/israel-hamas-war-spotlight-shaking-the-rust-off-sysjoker/

Emulated the Sysjoker C2 that was observed targeting Israel in the write-up by checkpoint.  The functions of the C2 is simple command execution and downloading files that are hosted in the static folder of where the server is executed.

[C2 Server](/projects/c2/c2server/main.go) - Central Server for the clients and operators

[C2 Client](/projects/c2/c2client/main.go) - This is the payload that is generated and allows for the asynchronous connect back to the c2 server.  The way the payload is designed makes it depend on a scheduled task to execute it or another process to run it periodically

[C2 Operator](/projects/c2/c2operator/main.go) - This is the red team operators client to communicate with the server and command the clients that are connected

![Whiteboard](/projects/c2/whiteboard.jpg)