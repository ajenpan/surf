package main

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"sync"
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
	wg := &sync.WaitGroup{}
	for cnt := 0; cnt < 100; cnt++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				work(remoteAddr)
			}
		}()
	}
	fmt.Println("wait.")

	wg.Wait()

	fmt.Println("work finished")

}

func work(remoteaddr string) {
	conn, err := net.DialTimeout("tcp", remoteaddr, 100*time.Millisecond)
	if err != nil {
		fmt.Printf("conn err:%v\n", err)
		return
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(15 * time.Second))

	l := rand.Int31n(math.MaxInt16)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()

		data := make([]byte, l)
		_, err = rand.Read(data)
		if err != nil {
			// fmt.Println(err)
			return
		}

		_, err = conn.Write(data)
		if err != nil {
			// fmt.Printf("Write err:%v\n", err)
			return
		}

	}()

	go func() {
		defer wg.Done()
		data := make([]byte, l)
		_, err = conn.Read(data)
		if err != nil {
			// fmt.Printf("Read err:%v\n", err)
			return
		}
	}()

	wg.Wait()

	fmt.Println("pass")
}
