package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	args := ParseArgs()

	start := time.Now()
	fmt.Printf("[%v] Scan started\n", start.Format("2006-01-02 15:04:05"))

	for ip := args.StartIp; bytes.Compare(ip, args.EndIp) < 1; ip = getNextIP(ip) {
		scan(ip, args.Ports)
	}

	end := time.Now()
	fmt.Printf("[%v] Scan complete\nDuration:%v\n", end.Format("2006-01-02 15:04:05"), end.Sub(start))
}

func getNextIP(cur net.IP) net.IP {
	//from least to most significant byte
	for i := len(cur) - 1; i >= 0; i-- {
		cur[i]++

		//if overflow then increment the next significant byte
		if cur[i] != 0 {
			return cur
		}
	}

	return cur
}

func scan(ip net.IP, ports []uint16) {
	var results []chan ConnectResult
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
	fmt.Printf("%v complete in %0f seconds\n", ip, stop.Sub((start)).Seconds())
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
