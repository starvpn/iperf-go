package iperf

import (
	"errors"
	"net"
	"strconv"
	"time"
)

type UDPProto struct {
}

func (u *UDPProto) name() string {
	return UDP_NAME
}

func (u *UDPProto) accept(test *IperfTest) (net.Conn, error) {
	Log.Debugf("Enter UDP accept")

	// UDP 不需要 accept，直接返回已创建的连接
	// 这里简单返回错误，实际使用中需要根据具体需求实现
	return nil, errors.New("UDP accept not implemented")
}

func (u *UDPProto) listen(test *IperfTest) (net.Listener, error) {
	Log.Debugf("Enter UDP listen")

	// UDP 不使用 net.Listener，返回 nil
	return nil, nil
}

func (u *UDPProto) connect(test *IperfTest) (net.Conn, error) {
	Log.Debugf("Enter UDP connect")

	udpAddr, err := net.ResolveUDPAddr("udp4", test.addr+":"+strconv.Itoa(int(test.port)))
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}

	err = conn.SetDeadline(time.Now().Add(time.Duration(test.duration+5) * time.Second))
	if err != nil {
		Log.Errorf("SetDeadline err: %v", err)
		return nil, err
	}

	return conn, nil
}

func (u *UDPProto) send(sp *iperfStream) int {
	n, err := sp.conn.(*net.UDPConn).Write(sp.buffer)
	if err != nil {
		Log.Errorf("udp write err = %T %v", err, err)
		return -2
	}

	if n < 0 {
		return n
	}

	sp.result.bytes_sent += uint64(n)
	sp.result.bytes_sent_this_interval += uint64(n)

	Log.Debugf("UDP sent %d bytes, total sent: %d", n, sp.result.bytes_sent)

	return n
}

func (u *UDPProto) recv(sp *iperfStream) int {
	n, err := sp.conn.(*net.UDPConn).Read(sp.buffer)
	if err != nil {
		Log.Errorf("udp recv err = %T %v", err, err)
		return -2
	}

	if n < 0 {
		return n
	}

	sp.result.bytes_received += uint64(n)
	sp.result.bytes_received_this_interval += uint64(n)

	Log.Debugf("UDP recv %d bytes, total recv: %d", n, sp.result.bytes_received)

	return n
}

func (u *UDPProto) init(test *IperfTest) int {
	Log.Debugf("Enter UDP init")

	// UDP 特定的初始化
	// 可以在这里设置 UDP 特定的参数

	return 0
}

func (u *UDPProto) teardown(test *IperfTest) int {
	Log.Debugf("Enter UDP teardown")

	// UDP 清理工作

	return 0
}

func (u *UDPProto) statsCallback(test *IperfTest, sp *iperfStream, tempResult *iperf_interval_results) int {
	// UDP 统计回调
	// 可以在这里收集 UDP 特定的统计信息，如丢包率等

	return 0
}
