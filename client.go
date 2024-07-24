package connecty

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net"
)

func CreateConnection(proxyUrl string, targetHost string, targetPort int) (net.Conn, error) {
	req := Request{
		Host: targetHost,
		Port: targetPort,
	}

	conn, err := websocket.Dial(proxyUrl, "", "http://localhost/")
	if err != nil {
		return nil, fmt.Errorf("failed to open ws: %v", err)
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode connection request: %v", err)
	}

	conn.Write(buf.Bytes())

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read connection request response: %v %v", err, resp)
	}

	log.Printf("received con request response: %v", resp)
	if resp.Status != "ok" {
		return nil, errors.New(resp.Error)
	}

	return conn, nil
}
