package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cpg1111/dhcpdebug/internal/client"
)

type cmdArgs struct {
	Iface       string
	ClientHost  string
	ClientPort  int
	Host        string
	Port        int
	BrdcstAddr  string
	Proto       int
	NumMessages int
	Release     bool
	Timeout     int
	Debug       bool
}

func run(args cmdArgs) int {
	var (
		c   client.Client
		err error
	)

	if args.Proto == 4 {
		c, err = client.NewDHCP4(
			args.Iface,
			args.ClientHost,
			args.Host,
			args.ClientPort,
			args.Port,
			args.Timeout,
			args.Debug,
		)
	} else {
		c, err = client.NewDHCP6(
			args.Iface,
			args.BrdcstAddr,
			args.Timeout,
			args.NumMessages,
			args.Debug,
		)
	}

	if err != nil {
		fmt.Println(err)
		return 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go c.Exec(ctx, args.Release)

	select {
	case <-sigChan:
		return 0
	case <-c.Done():
		err = c.Err()
		if err != nil {
			fmt.Println(err)
			return 1
		}

		return 0
	}
}

func main() {
	var args cmdArgs

	flag.StringVar(&args.Iface, "iface", "", "interface to send broadcast packets over")
	flag.StringVar(&args.Host, "client-host", "", "unicast source address")
	flag.IntVar(&args.Port, "client-port", 0, "unicast source port")
	flag.StringVar(&args.Host, "host", "", "destination server host")
	flag.IntVar(&args.Port, "port", 67, "destination server port")
	flag.StringVar(&args.BrdcstAddr, "broadcast-addr", "", "dhcp6 broadcast address")
	flag.IntVar(&args.Proto, "proto", 4, "DHCP protocol version, either 4 or 6")
	flag.IntVar(&args.NumMessages, "num-msg", 4, "Number of messages to use for DHCP6")
	flag.BoolVar(&args.Release, "release", false, "whether to immediately release IP")
	flag.IntVar(&args.Timeout, "timeout", 30000, "overall timeout in milliseconds")
	flag.BoolVar(&args.Debug, "debug", false, "debug mode")
	flag.Parse()

	os.Exit(run(args))
}
