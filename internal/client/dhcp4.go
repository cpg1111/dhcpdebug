package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4/nclient4"
)

const (
	DiscoverTimeout = 15 * time.Second
	RequestTimeout  = 15 * time.Second
)

type DHCP4 struct {
	iface   *net.Interface
	client  *nclient4.Client
	timeout time.Duration
	done    chan struct{}
	err     error
}

func NewDHCP4(iface, srcHost, dstHost string, srcPort, dstPort, timeout int, debug bool) (*DHCP4, error) {
	d := &DHCP4{
		timeout: time.Duration(timeout) * time.Millisecond,
		done:    make(chan struct{}),
	}

	realIface, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}

	d.iface = realIface

	var opts []nclient4.ClientOpt

	if debug {
		opts = append(opts, nclient4.WithDebugLogger())
	}

	if srcHost != "" {
		srcAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", srcHost, srcPort))
		if err != nil {
			return nil, err
		}

		opts = append(opts, nclient4.WithUnicast(srcAddr))
	}

	if dstHost != "" && dstPort > 0 {
		dstAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", dstHost, dstPort))
		if err != nil {
			return nil, err
		}

		opts = append(opts, nclient4.WithServerAddr(dstAddr))
	}

	d.client, err = nclient4.New(iface, opts...)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DHCP4) Done() <-chan struct{} {
	return d.done
}

func (d *DHCP4) Err() error {
	return d.err
}

func (d *DHCP4) dora(ctx context.Context, release bool) error {
	discCtx, dCancel := context.WithTimeout(ctx, DiscoverTimeout)
	defer dCancel()

	offer, err := d.client.DiscoverOffer(discCtx)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", offer)

	reqCtx, rCancel := context.WithTimeout(ctx, RequestTimeout)
	defer rCancel()

	lease, err := d.client.RequestFromOffer(reqCtx, offer)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", lease)

	if release {
		d.client.Release(lease)
	}

	return nil
}

func (d *DHCP4) Exec(ctx context.Context, release bool) {
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

	d.err = d.dora(ctx, release)
}
