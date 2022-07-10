package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type AppArgs struct {
	Ports   []uint16
	StartIp net.IP
	EndIp   net.IP
}

func ParseArgs() *AppArgs {
	startIp := flag.String("sip", "", "Start of ip range, must be less than or equal to eip. No wild cards supported")
	endIp := flag.String("eip", "", "End of ip range, must be greater than or equal to sip. No wild cards supported")
	userPorts := flag.String("p", "80", "Ports to scan eg `-p 80,443`. If not provided all ports will be scanned.")
	flag.Parse()

	appArgs := AppArgs{}

	if *startIp != "" {
		appArgs.StartIp = net.ParseIP(*startIp)
	} else {
		appArgs.StartIp = net.IPv4zero
	}

	if *endIp != "" {
		appArgs.EndIp = net.ParseIP(*endIp)
	} else {
		appArgs.EndIp = net.IPv4bcast
	}

	if userPorts != nil {
		ports := strings.Split(*userPorts, ",")
		for _, portStr := range ports {
			port, _ := strconv.Atoi(portStr)
			appArgs.Ports = append(appArgs.Ports, uint16(port))
		}
	}

	fmt.Printf("Ip range:(%v) - (%v)\n", appArgs.StartIp, appArgs.EndIp)
	fmt.Printf("Ports: %v\n", *userPorts)

	return &appArgs
}
