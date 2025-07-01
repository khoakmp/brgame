package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

func runTCPServer() {
	l, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Println(err)
		return
	}

	//unix.Pipe()
	serveFn := func(conn net.Conn) {

		for {
			var buf [1024]byte
			time.Sleep(time.Second)

			n, err := conn.Write(buf[:])

			if err != nil {
				fmt.Println("Failed to write:", err)
				conn.Close()
				return
			}

			fmt.Println("Write ", n, "bytes successfully")

			n, err = conn.Read(buf[:])
			if err != nil {
				fmt.Println("Failed to read from conn:", err)
				conn.Close()
				return
			}
			fmt.Println("read:", n)
		}
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept conn", err)
			return
		}
		tcpConn := conn.(*net.TCPConn)
		fmt.Println("[SERVER] Accept new conn, remote addr:", tcpConn.RemoteAddr().String(), "local addr:", tcpConn.LocalAddr().String())
		rc, _ := tcpConn.SyscallConn()
		rc.Control(func(fd uintptr) {

			sndbuf, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF)
			if err != nil {
				fmt.Println("[SERVER] failed to get snd buf size:", err)
			}
			fmt.Println("[CLIENT] snd buf size:", sndbuf)

		})
		go serveFn(conn)
	}

}

func runTCPClient() {
	numClients := 10
	var wg sync.WaitGroup
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 10)
			conn, err := net.Dial("tcp", ":8082")
			if err != nil {
				fmt.Println("failed to connect listener:", err)
				return
			}
			defer conn.Close()
			tcpConn := conn.(*net.TCPConn)
			rc, err := tcpConn.SyscallConn()
			if err != nil {
				fmt.Println("[CLIENT] failed to get raw conn, ", err)
				return
			}

			rc.Control(func(fd uintptr) {
				sndBufSize, err := unix.GetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_SNDBUF)
				if err != nil {
					fmt.Println("[CLIENT] failed to get snd buf size:", err)
					return
				}
				fmt.Println("[CLIENT] snd buf size:", sndBufSize)
			})

			//fmt.Println("[CLIENT] remote addr:", tcpConn.RemoteAddr().String(), "local addr:", tcpConn.LocalAddr().String())
			//fmt.Println("Sleep 2 sec...")
			time.Sleep(time.Second * 2)
			//fmt.Println("Closing")
		}()
	}

	wg.Wait()
}
func ExpTCP() {
	cmd := os.Args[1]
	switch cmd {
	case "s":
		runTCPServer()
	case "c":
		runTCPClient()
	}
}
