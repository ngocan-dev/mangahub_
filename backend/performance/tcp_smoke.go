package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:9000", "TCP address to test")
	clients := flag.Int("clients", 50, "number of concurrent clients")
	hold := flag.Duration("hold", 2*time.Second, "how long to keep each connection open")
	timeout := flag.Duration("timeout", 1*time.Second, "dial timeout")
	flag.Parse()

	var success uint64
	var failure uint64

	var wg sync.WaitGroup
	wg.Add(*clients)

	for i := 0; i < *clients; i++ {
		go func(id int) {
			defer wg.Done()

			d := net.Dialer{Timeout: *timeout}
			conn, err := d.Dial("tcp", *addr)
			if err != nil {
				log.Printf("dial %d failed: %v", id, err)
				atomic.AddUint64(&failure, 1)
				return
			}
			defer conn.Close()

			// Attempt a lightweight write to exercise the socket
			if _, err := conn.Write([]byte("ping\n")); err != nil {
				log.Printf("dial %d write failed: %v", id, err)
			}

			atomic.AddUint64(&success, 1)

			// Keep the connection alive briefly to emulate a streaming client
			conn.SetDeadline(time.Now().Add(*hold))
			time.Sleep(*hold)
		}(i)
	}

	wg.Wait()

	total := uint64(*clients)
	fmt.Printf("connections attempted: %d\n", total)
	fmt.Printf("successful: %d\n", atomic.LoadUint64(&success))
	fmt.Printf("failed: %d\n", atomic.LoadUint64(&failure))

	successRate := float64(success) / float64(total)
	fmt.Printf("success rate: %.2f%%\n", successRate*100)

	if successRate < 0.99 {
		log.Fatalf("TCP connection success rate below 99%%")
	}
}
