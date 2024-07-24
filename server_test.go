package connecty

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"testing"
	"time"
)

func startTestServer() {
	http.Handle("/connect", websocket.Handler(HandleWebsocket))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func TestSSHProxy(t *testing.T) {
	go startTestServer()

	// Wait for the server to start
	time.Sleep(3 * time.Second)

	conn, err := CreateConnection("ws://localhost:8080/connect", "localhost", 2222)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}

	if conn == nil {
		t.Fatalf("connection is nil")
	}

	defer conn.Close()

	sshClientConfig := ssh.ClientConfig{
		User: "test",
		Auth: []ssh.AuthMethod{
			ssh.Password("test"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cc, nc, r, err := ssh.NewClientConn(conn, "127.0.0.1", &sshClientConfig)
	if err != nil {
		t.Fatalf("failed to create ssh client connection: %v", err)
	}

	client := ssh.NewClient(cc, nc, r)
	if client == nil {
		t.Fatalf("client is nil")
	} else {
		defer client.Close()
	}

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	} else {
		defer session.Close()
	}

	out, err := session.Output("ls")
	if err != nil {
		t.Fatalf("failed to run command: %v", err)
	}

	t.Logf("output: %s", out)
}
