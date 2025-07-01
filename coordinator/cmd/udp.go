package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func ExpUdp() {
	var wg sync.WaitGroup
	n := 5
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			addr, _ := net.ResolveUDPAddr("udp", ":5678")
			listener, err := net.ListenUDP("udp", addr)
			if err != nil {
				fmt.Println("failed to listen udp at port 5678", err)
				return
			}

			fmt.Println(listener.LocalAddr().String())
			time.Sleep(time.Second)
			listener.Close()
		}()
	}

	wg.Wait()
}
