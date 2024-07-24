package connecty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"golang.org/x/net/websocket"
)

func ping(ws *websocket.Conn, done chan struct{}) {
	w, err := ws.NewFrameWriter(websocket.PingFrame)
	if err != nil {
		log.Printf("failed to create pingwriter: %v", err)
		return
	}

	ticker := time.Tick(20 * time.Second)
	for {
		select {
		case <-ticker:
			_, err = w.Write(nil)
			if err != nil {
				log.Printf("failed to write ping msg: %v", err)
				return
			}
		case <-done:
			return
		}
	}
}

func HandleWebsocket(wsConn *websocket.Conn) {
	// Read the connection request
	err := wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Printf("failed to set read deadline: %v", err)
		return
	}
	buf := make([]byte, 2048)
	_, err = wsConn.Read(buf)
	if err != nil {
		log.Printf("failed to read connection request: %v", err)
		return
	}

	var req Request
	err = json.NewDecoder(bytes.NewBuffer(buf)).Decode(&req)
	if err != nil {
		log.Printf("failed to parse connection request [%s]: %v", buf, err)
		return
	}

	// Clear the deadline
	err = wsConn.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("failed to clear connection deadline: %v", err)
		return
	}

	log.Printf("connecting to %s on port %d", req.Host, req.Port)

	// Connect to the remote host
	var resp Response

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", req.Host, req.Port), 30*time.Second)
	if err != nil {
		log.Printf("failed to connect: %v", err)

		resp.Status = "failed"
		resp.Error = err.Error()
		r, err := json.Marshal(resp)
		if err != nil {
			log.Printf("failed to marshall response: %v", err)
		} else {
			if err := websocket.Message.Send(wsConn, r); err != nil {
				log.Printf("failed to write response: %v", err)
			}
		}

		return
	}
	defer conn.Close()

	// Send the status back to the client
	resp.Status = "ok"
	r, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshall response: %v", err)
	} else {
		if err := websocket.Message.Send(wsConn, r); err != nil {
			log.Printf("failed to write response: %v", err)
		}
	}

	wsConn.PayloadType = websocket.BinaryFrame

	// Start the proxy
	done := make(chan struct{})

	// Start the ping
	go ping(wsConn, done)

	// Proxy the connection
	status := make(chan connectionStatus)

	// conn -> wsConn
	go func() {
		b, err := io.Copy(conn, wsConn)
		if err != nil {
			log.Printf("failed to copy ws to conn: %v", err)
		}

		conn.Close()
		status <- connectionStatus{"up", err, b}
	}()

	// wsConn -> conn
	go func() {
		b, err := io.Copy(wsConn, conn)
		if err != nil {
			log.Printf("failed to copy conn to ws: %v", err)
		}

		wsConn.Close()
		status <- connectionStatus{"down", err, b}
	}()

	// Wait for the connection to close
	s := <-status
	log.Printf("connection closed: %s, %d bytes, %v", s.dir, s.bytes, s.err)

	close(done)
}
