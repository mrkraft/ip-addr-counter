package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	filePath := flag.String("f", "ip_addresses", "file path") // .\ip-addr-counter.exe -f="ip_addresses"
	flag.Parse()

	execStartTime := time.Now()
	fmt.Printf("Start time: %s\n", execStartTime.Local())

	resultCount, err := getUniqueIpv4AddrNumber(*filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	execDuration := time.Since(execStartTime)

	fmt.Printf("Number of unique IPv4: %d\n", resultCount)
	fmt.Printf("Time: %s\n", execDuration)
}
