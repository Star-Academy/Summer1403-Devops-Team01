package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	ProtocolICMP = 1
	MAXTTL       = 128
)

var ctx = context.Background()
var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func main() {
	router := gin.Default()
	router.GET("/traceroute/:host", trace)

	router.Run("localhost:8080")
}

func trace(c *gin.Context) {
	host := c.Param("host")

	// Resolve IP address
	ipAddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		fmt.Println("Error resolving IP address:", err)
		c.IndentedJSON(
			http.StatusInternalServerError,
			gin.H{"ERROR": "Failed to resolve IP address"},
		)
		return
	}

	// traceResponse := make(map[int]string)
	var traceResponse []struct {
		Key   int
		Value string
	}

	for ttl := 1; ttl <= MAXTTL; ttl++ {
		// Create a raw socket
		conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			fmt.Println("Error creating socket:", err)
			c.IndentedJSON(
				http.StatusInternalServerError,
				gin.H{"ERROR": "Failed to create socket"},
			)
			return
		}
		defer conn.Close()

		if err := conn.IPv4PacketConn().SetTTL(ttl); err != nil {
			fmt.Println("Error setting TTL:", err)
			c.IndentedJSON(
				http.StatusInternalServerError,
				gin.H{"ERROR": "Failed to set ttl"},
			)
			return
		}

		// Create ICMP packet
		icmpMsg := make([]byte, 8)
		icmpMsg[0] = 8
		icmpMsg[1] = 0
		icmpMsg[2] = 0
		icmpMsg[3] = 0
		icmpMsg[4] = 0
		icmpMsg[5] = 1
		icmpMsg[6] = 0
		icmpMsg[7] = 2

		checksum := checksum(icmpMsg)
		icmpMsg[2] = byte(checksum >> 8)
		icmpMsg[3] = byte(checksum)

		// msg := icmp.Message {
		// 	Type: ipv4.ICMPTypeEcho, Code: 0,
		// 	Body: &icmp.Echo{
		// 		ID: os.Getpid() & 0xffff, Seq: 1,
		// 		Data: []byte(""),
		// 	},
		// }
		// icmpMsg, err := msg.Marshal(nil)
		// if err != nil {
		//     fmt.Println("Error creating icmpMsg: ", err)
		//     return
		// }

		// Send ICMP packet
		start := time.Now()

		_, err = conn.WriteTo(icmpMsg, ipAddr)
		if err != nil {
			fmt.Println(err)
			c.IndentedJSON(
				http.StatusInternalServerError,
				gin.H{"ERROR": "Failed to write"},
			)
			return
		}

		buff := make([]byte, 512)
		err = conn.SetReadDeadline(time.Now().Add(250 * time.Millisecond))
		if err != nil {
			fmt.Println("set read deadline error: ", err)
			c.IndentedJSON(
				http.StatusInternalServerError,
				gin.H{"ERROR": "Failed to set read deadline"},
			)
			return
		}

		n, addr, err := conn.ReadFrom(buff)
		if err != nil {
			// fmt.Println("Read error: ", err)
			fmt.Println("*\t*\t*")
			// traceResponse[ttl] = "Timed Out"
			traceResponse = append(traceResponse, struct {
				Key   int
				Value string
			}{Key: ttl, Value: "Time Out."})
			continue
		}

		duration := time.Since(start)

		rm, err := icmp.ParseMessage(ProtocolICMP, buff[:n])
		if err != nil {
			fmt.Println(err)
			c.IndentedJSON(
				http.StatusInternalServerError,
				gin.H{"ERROR": "Failed to parse ICMP message"},
			)
			return
		}

		switch rm.Type {

		case ipv4.ICMPTypeEchoReply:
			// traceResponse[ttl] = ipAddr.String() + duration.String()
			traceResponse = append(traceResponse, struct {
				Key   int
				Value string
			}{Key: ttl, Value: ipAddr.String() + "\t" + duration.String()})

			fmt.Println(ipAddr, ttl, duration)
			c.IndentedJSON(
				http.StatusOK,
				traceResponse,
			)

			sort.Slice(traceResponse, func(i, j int) bool {
				return traceResponse[i].Key < traceResponse[j].Key
			})

			jsondata, _ := json.Marshal(traceResponse)

			err := client.Set(ctx, host+start.String(), jsondata, 0).Err()
			if err != nil {
				log.Fatalf("could not set hash: %v", err)
			}

			val, err := client.Get(ctx, host+start.String()).Result()
			if err != nil {
				log.Fatalf("could not get hash: %v", err)
			}

			var rmap map[int]string

			err = json.Unmarshal([]byte(val), &rmap)

			fmt.Println(rmap)
			return

		case ipv4.ICMPTypeTimeExceeded:
			// traceResponse[ttl] = addr.String() + duration.String()
			traceResponse = append(traceResponse, struct {
				Key   int
				Value string
			}{Key: ttl, Value: addr.String() + "\t" + duration.String()})

			fmt.Println(&net.IPAddr{IP: addr.(*net.IPAddr).IP}, ttl, duration)

		default:
			fmt.Printf("got %+v from %v; want echo reply", rm, addr)
		}
	}
	c.IndentedJSON(
		http.StatusOK,
		traceResponse,
	)
}

func checksum(msg []byte) uint16 {
	sum := 0
	for i := 0; i < len(msg)-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if len(msg)%2 == 1 {
		sum += int(msg[len(msg)-1]) * 256
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	return uint16(^sum)
}
