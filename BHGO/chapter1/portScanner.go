package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	ip := "10.17.37.58"
	for i := 1; i <= 65534; i++ {
		port := strconv.Itoa(i)
		connString := ip + ":" + port
		//fmt.Println(connString)
		conn, err := net.Dial("tcp", connString)
		if err == nil {
			fmt.Printf("Connection Successful %d\n", i)
		} else {
			continue
		}
		conn.Close()
	}
}
