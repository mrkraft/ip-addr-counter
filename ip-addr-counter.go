package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	subIpv4Num           = 2097152 // for IPv4 a.b.c.d - max number of b.c.d / 8 = 2097152 bites
	workersNum           = 32      // Number of workers for parallel processing
	ipListChannelBufSize = 1000
	fileReaderBufSize    = 256 * 1024
)

var (
	counter                 atomic.Uint64 // unique IPv4 counter
	ipAddrMatrixColumnLocks []sync.Mutex
)

// Converts last 3 bytes from IPv4 to int
func parseIPv4ToInt(ipv4 net.IP) (uint32, error) {
	return uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3]), nil
}

// Checks unique IPs and update the counter
func checkAndSetState(uniqueIpAddrState2D [][]byte, ipv4 net.IP) {
	ipInt, err := parseIPv4ToInt(ipv4)
	if err != nil {
		return // Skip invalid IP addresses
	}

	byteIndex := ipInt / 8
	bitIndex := ipInt % 8

	ipAddrMatrixColumnLocks[ipv4[0]].Lock()
	if uniqueIpAddrState2D[ipv4[0]][byteIndex]&(1<<bitIndex) == 0 {
		counter.Add(1)
		uniqueIpAddrState2D[ipv4[0]][byteIndex] |= 1 << bitIndex
	}
	ipAddrMatrixColumnLocks[ipv4[0]].Unlock()
}

// worker function that processes IP addresses from the input channel
func worker(uniqueIpAddrState2D [][]byte, ipChannel <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()

	for ipListBuf := range ipChannel {
		stringAll := string(ipListBuf)
		stringAll = strings.TrimSuffix(stringAll, "\n")
		ipStrList := strings.Split(stringAll, "\n")

		for _, ipStr := range ipStrList {
			ip := net.ParseIP(ipStr)

			if ip == nil || ip.To4() == nil {
				fmt.Printf("%s - invalid or non-IPv4 address, skip it\n", ipStr)
			}

			ipv4 := ip.To4()

			checkAndSetState(uniqueIpAddrState2D, ipv4)
		}
	}
}

// Processes a file and counts unique IPs in parallel
func getUniqueIpv4AddrNumber(filePath string) (uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Create matrix (256 X 2097152 bytes) to track unique IPs
	uniqueIpAddrState2D := make([][]byte, 256)
	for i := range uniqueIpAddrState2D {
		uniqueIpAddrState2D[i] = make([]byte, subIpv4Num)
	}

	// Create 256 mutexes for each column of the matrix
	// to allow multiple threads access it without locking (if they access different columns)
	for i := 0; i < 256; i++ {
		ipAddrMatrixColumnLocks = append(ipAddrMatrixColumnLocks, sync.Mutex{})
	}

	// Channel for distributing IPs to workers
	ipListChannel := make(chan []byte, ipListChannelBufSize)

	var wg sync.WaitGroup

	for i := 0; i < workersNum; i++ {
		wg.Add(1)
		go worker(uniqueIpAddrState2D, ipListChannel, &wg)
	}

	// Read the file
	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, fileReaderBufSize)
		return lines
	}}

	for {
		buf := linesPool.Get().([]byte)
		n, err := io.ReadFull(reader, buf)
		buf = buf[:n]

		if n == 0 {
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				break
			}
		}

		nextUntilNewline, err := reader.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntilNewline...)
		}

		ipListChannel <- buf
	}

	// Close the ipListChannel and wait for workers to finish
	close(ipListChannel)
	wg.Wait()

	return counter.Load(), nil
}
