package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/dhcpv6/nclient6"
)

const (
	SolicitTimeout  = 15 * time.Second
	Request6Timeout = 15 * time.Second
)

type DHCP6 struct {
	iface   *net.Interface
	client  *nclient6.Client
	timeout time.Duration
	done    chan struct{}
	err     error
	numMsgs int
}

func NewDHCP6(iface, brdcstAddr string, timeout, numMessages int, debug bool) (*DHCP6, error) {
	d := &DHCP6{
		timeout: time.Duration(timeout) * time.Millisecond,
		done:    make(chan struct{}),
		numMsgs: numMessages,
	}

	realIface, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}

	d.iface = realIface

	var opts []nclient6.ClientOpt

	if brdcstAddr != "" {
		addr, err := net.ResolveUDPAddr("udp6", net.JoinHostPort(brdcstAddr, "67"))
		if err != nil {
			return nil, err
		}

		opts = append(opts, nclient6.WithBroadcastAddr(addr))
	}

	if debug {
		opts = append(opts, nclient6.WithDebugLogger())
	}

	d.client, err = nclient6.New(iface, opts...)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DHCP6) Done() <-chan struct{} {
	return d.done
}

func (d *DHCP6) Err() error {
	return d.err
}

func (d *DHCP6) exchange(ctx context.Context, release bool) error {
	sCtx, sCancel := context.WithTimeout(ctx, SolicitTimeout)
	defer sCancel()

	var (
		adv, reply *dhcpv6.Message
		err        error
	)

	if d.numMsgs == 4 {
		adv, err = d.client.Solicit(sCtx)
		if err != nil {
			return err
		}

		fmt.Printf("%+v\n", adv)
		fmt.Printf("%+v\n", adv.Options)

		rCtx, rCancel := context.WithTimeout(ctx, Request6Timeout)
		defer rCancel()

		reply, err = d.client.Request(rCtx, adv)
	} else {
		reply, err = d.client.RapidSolicit(sCtx)
	}

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", reply)
	return nil
}

func (d *DHCP6) Exec(ctx context.Context, release bool) {
	defer func() {
		d.done <- struct{}{}
	}()

	defer func() {
		err := d.client.Close()
		if err != nil {
			fmt.Println(err)
			if d.err == nil {
				d.err = err
			}
		}
	}()

	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(ctx, d.timeout)
	defer cancel()

	d.err = d.exchange(ctx, release)
}
