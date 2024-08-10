package trace

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ProtocolICMP = 1
	MAXTTL       = 128
	TIMEOUT_TIME = 250
)

type TraceResponse struct {
	TTL   int
	IP    string
	RTIME string
}

func ResolveIP(host string) (*net.IPAddr, error) {
	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		fmt.Println("Error resolving IP address:", err)
	}
	return ipAddr, err
}

func PerformTrace(ipAddr *net.IPAddr) ([]TraceResponse, error) {
	var traceResponses []TraceResponse

	for ttl := 1; ttl <= MAXTTL; ttl++ {
		conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			fmt.Println("Error creating socket:", err)
			return nil, fmt.Errorf("error creating socket: %w", err)
		}
		defer conn.Close()

		if err := conn.IPv4PacketConn().SetTTL(ttl); err != nil {
			fmt.Println("Error setting TTL:", err)
			return nil, fmt.Errorf("error setting TTL: %w", err)
		}

		msg := createICMPMessage()
		start := time.Now()

		if err := sendICMPMessage(conn, msg, ipAddr); err != nil {
			fmt.Println("Error sending ICMP message:", err)
			return nil, fmt.Errorf("error sending ICMP message: %w", err)
		}

		curResponse := receiveICMPResponse(conn, ttl, start)
		traceResponses = append(traceResponses, curResponse)
		if curResponse.IP == ipAddr.String() {
			break
		}
	}
	return traceResponses, nil
}

func createICMPMessage() icmp.Message {
	return icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
}

func sendICMPMessage(conn *icmp.PacketConn, msg icmp.Message, ipAddr *net.IPAddr) error {
	b, err := msg.Marshal(nil)
	if err != nil {
		fmt.Println("Error marshalling ICMP message:", err)
		return err
	}

	if _, err := conn.WriteTo(b, ipAddr); err != nil {
		fmt.Println("Error writing ICMP message:", err)
		return err
	}
	return nil
}

func receiveICMPResponse(conn *icmp.PacketConn, ttl int, start time.Time) TraceResponse {
	buff := make([]byte, 1500)

	err := conn.SetReadDeadline(time.Now().Add(TIMEOUT_TIME * time.Millisecond))
	if err != nil {
		fmt.Println("Error setting ReadDeadline:", err)
		return TraceResponse{TTL: ttl, IP: "*", RTIME: "Set Deadline error!"}
	}

	n, addr, err := conn.ReadFrom(buff)
	if err != nil {
		fmt.Println("*\t*\t*")
		return TraceResponse{TTL: ttl, IP: "*", RTIME: "Time out!"}
	}

	duration := time.Since(start)

	rm, err := icmp.ParseMessage(ProtocolICMP, buff[:n])
	if err != nil {
		fmt.Println("Error parsing ICMP message:", err)
		return TraceResponse{TTL: ttl, IP: "*", RTIME: "Parse error"}
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return TraceResponse{TTL: ttl, IP: addr.String(), RTIME: duration.String()}
	case ipv4.ICMPTypeTimeExceeded:
		return TraceResponse{TTL: ttl, IP: addr.String(), RTIME: duration.String()}
	default:
		fmt.Printf("got %+v from %v; want echo reply", rm, addr)
		return TraceResponse{TTL: ttl, IP: addr.String(), RTIME: "Unknown response"}
	}
}
