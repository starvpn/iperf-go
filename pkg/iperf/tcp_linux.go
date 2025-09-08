//go:build linux
// +build linux

package iperf

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

func saveTCPInfo(sp *iperfStream, rp *iperf_interval_results) int {
	info := getTCPInfo(sp.conn)

	rp.rtt = uint(info.Rtt)
	rp.rto = uint(info.Rto)
	rp.interval_retrans = uint(info.Total_retrans)

	return 0
}

func getTCPInfo(conn net.Conn) *unix.TCPInfo {
	file, err := conn.(*net.TCPConn).File()
	if err != nil {
		fmt.Printf("File err: %v\n", err)

		return &unix.TCPInfo{}
	}

	defer file.Close()

	fd := file.Fd()

	info, err := unix.GetsockoptTCPInfo(int(fd), unix.SOL_TCP, unix.TCP_INFO)
	if err != nil {
		fmt.Printf("GetsockoptTCPInfo err: %v\n", err)

		return &unix.TCPInfo{}
	}

	return info
}

func PrintTCPInfo(info *unix.TCPInfo) {
	fmt.Printf("TcpInfo: rcv_rtt:%v\trtt:%v\tretransmits:%v\trto:%v\tlost:%v\tretrans:%v\ttotal_retrans:%v\n",
		info.Rcv_rtt,
		info.Rtt,
		info.Retransmits,
		info.Rto,
		info.Lost,
		info.Retrans,
		info.Total_retrans,
	)
}
