package main

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	var remoteAddr string
	if len(os.Args) <= 1 {
		remoteAddr = "101.133.149.209:9998"
	} else {
		remoteAddr = os.Args[1]
	}

	fmt.Println("start to work")

	for i := 0; i < 10000; i++ {
		if i%100 == 0 {
			fmt.Print(i, "-")
		}
		work(remoteAddr)
		if i%100 == 0 {
			fmt.Println(i)
		}
	}
}

func work(remoteaddr string) {

	// net.DialTCP("tcp",remoteaddr)
	conn, err := net.DialTimeout("tcp", remoteaddr, 100*time.Microsecond)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	l := rand.Int31n(math.MaxInt16)
	data := make([]byte, l)
	_, err = rand.Read(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		return
	}

	_, err = conn.Read(data)
	if err != nil {
		return
	}
	fmt.Println("pass")
}
