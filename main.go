package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	//todo: read ports from -p --ports arguments
	ports := [15]uint16{80, 81, 443, 4550, 5160, 5511, 5550, 5554, 8080, 6550, 7000, 8000, 8866, 56000, 10000}
	var results []chan ConnectResult

	for ip := net.IPv4zero.To4(); !ip.Equal(net.IPv4bcast); ip = getNextIP(ip) {
		start := time.Now()
		//go routines to make requests on ports concurrently
		for _, port := range ports {
			results = append(results, canConnect(ip, port))
		}

		//https://go.dev/blog/pipelines
		//go routines to display results as they return from the host using the fanIn pattern
		for result := range merge(results) {
			if result.Ok {
				fmt.Printf("%v\tconnected\n", result.Host)
			}
		}
		stop := time.Now()
		fmt.Printf("%v complete in %v\n", ip, stop.Sub((start)).Seconds())
	}
}

type ConnectResult struct {
	Ok   bool
	Host string
}

func canConnect(ip net.IP, port uint16) chan ConnectResult {
	url := fmt.Sprintf("%v:%d", ip, port)
	out := make(chan ConnectResult)

	go func() {
		con, err := net.DialTimeout("tcp", url, time.Second*5)
		if err != nil {
			out <- ConnectResult{
				Ok:   false,
				Host: url,
			}
		} else {
			con.Close()
			out <- ConnectResult{
				Ok:   true,
				Host: url,
			}
		}
		close(out)
	}()

	return out
}

func merge(cs []chan ConnectResult) <-chan ConnectResult {
	var wg sync.WaitGroup
	out := make(chan ConnectResult)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan ConnectResult) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func getNextIP(cur net.IP) net.IP {
	//from least to most significant byte
	for i := len(cur) - 1; i >= 0; i-- {
		cur[i]++

		//if it overflowed then increment the next significant byte
		if cur[i] != 0 {
			return cur
		}
	}

	return cur
}
