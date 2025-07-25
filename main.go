package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

/*
	Usage:
	export MCP_ENDPOINT=<mcp_endpoint>

*/

// Load environment variables from .env file
// You can use a library like "github.com/joho/godotenv" to load .env file
var (
	MCP_ENDPOINT = os.Getenv("MCP_ENDPOINT")
)

// Reconnection settings
const (
	INITIAL_BACKOFF = 1   // Initial wait time in seconds
	MAX_BACKOFF     = 600 // Maximum wait time in seconds
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: xiaozhi-mcp-pipe <command> arg1 arg2 ...")
	}

	command := os.Args[1:]

	if MCP_ENDPOINT == "" {
		log.Fatalf("Please set the `MCP_ENDPOINT` environment variable")
	}

	// Register signal handler
	cleanup := make(chan os.Signal, 1)
	signal.Notify(cleanup, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-cleanup
		fmt.Println("Received interrupt signal, shutting down...")
		os.Exit(0)
	}()

	// Start main loop
	for {
		err := connectWithRetry(MCP_ENDPOINT, command)
		if err != nil {
			log.Printf("Program execution error: %v", err)
		}
	}
}

func connectWithRetry(uri string, parts []string) error {
	reconnectAttempt := 0
	backoff := INITIAL_BACKOFF

	for {
		if reconnectAttempt > 0 {
			waitTime := time.Duration(backoff) * time.Second
			log.Printf("Waiting %v seconds before reconnection attempt %d...", waitTime, reconnectAttempt)
			time.Sleep(waitTime)
		}

		err := connectToServer(uri, parts)
		if err == nil {
			return nil
		}

		reconnectAttempt++
		log.Printf("Connection closed (attempt: %d): %v", reconnectAttempt, err)

		// Calculate wait time for next reconnection (exponential backoff)
		backoff = min(backoff*2, MAX_BACKOFF)
	}
}

func connectToServer(uri string, parts []string) error {
	log.Printf("Connecting to WebSocket server...")

	// Establish WebSocket connection
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	conn, _, err := dialer.Dial(uri, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	log.Printf("Successfully connected to WebSocket server")

	// Start command process
	var cmd *exec.Cmd
	if len(parts) > 1 {
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		cmd = exec.Command(parts[0])
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	log.Printf("Started %s process", parts)

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all pipes
	go pipeWebSocketToProcess(ctx, conn, stdin)
	go pipeProcessToWebSocket(ctx, conn, stdout)
	go pipeProcessStderrToTerminal(ctx, stderr)

	// Wait for the process to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("Process exited with error: %v", err)
	}

	return nil
}

func pipeWebSocketToProcess(ctx context.Context, conn *websocket.Conn, stdin io.WriteCloser) {
	defer stdin.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error in WebSocket to process pipe: %v", err)
				return
			}
			log.Printf("<< %s", message)

			message = append(message, '\n')
			_, err = stdin.Write(message)
			if err != nil {
				log.Printf("Error writing to process stdin: %v", err)
				return
			}
		}
	}
}

func pipeProcessToWebSocket(ctx context.Context, conn *websocket.Conn, stdout io.Reader) {
	defer conn.Close()
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Bytes()
			log.Printf(">> %s", line)
			if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
				log.Printf("Error sending data to WebSocket: %v", err)
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from process stdout: %v", err)
	}
}

func pipeProcessStderrToTerminal(ctx context.Context, stderr io.Reader) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data := make([]byte, 1024)
			n, err := stderr.Read(data)
			if err != nil {
				log.Printf("Error reading from process stderr: %v", err)
				return
			}
			if n == 0 {
				log.Printf("Process has ended stderr output")
				return
			}

			os.Stderr.Write(data[:n])
		}
	}
}
