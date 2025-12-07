package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
)

// Session manages the interactive chat lifecycle.
type Session struct {
	client   *WSClient
	room     string
	verbose  bool
	quiet    bool
	stopChan chan struct{}
}

// NewSession initializes a chat session for the provided room.
func NewSession(room string) *Session {
	runtime := config.Runtime()
	return &Session{
		client:   NewWSClient(room),
		room:     room,
		verbose:  runtime.Verbose,
		quiet:    runtime.Quiet,
		stopChan: make(chan struct{}),
	}
}

// Run connects to the chat server and starts the interactive shell.
func (s *Session) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.client.Connect(ctx); err != nil {
		return err
	}

	s.printWelcome()

	go s.listen()
	go s.listenSignals()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			s.client.Close()
			return err
		}

		select {
		case <-s.stopChan:
			return nil
		default:
		}

		s.handleInput(strings.TrimSpace(line))
	}
}

func (s *Session) listenSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
	close(s.stopChan)
	_ = s.client.Close()
	fmt.Println("\nDisconnected from chat.")
	os.Exit(0)
}

func (s *Session) listen() {
	for {
		msg, err := s.client.Read()
		if err != nil {
			select {
			case <-s.stopChan:
				return
			default:
				fmt.Printf("✗ Connection closed: %v\n", err)
				close(s.stopChan)
				return
			}
		}

		s.renderMessage(msg)
	}
}

func (s *Session) renderMessage(msg Message) {
	prefix := msg.Timestamp.Local().Format("15:04")
	author := msg.User
	if author == "" {
		author = "system"
	}
	if s.quiet {
		fmt.Printf("[%s] %s\n", prefix, msg.Text)
		return
	}
	if msg.Type == "system" && msg.Text != "" {
		fmt.Printf("[%s] %s\n", prefix, msg.Text)
		return
	}

	fmt.Printf("[%s] %s: %s\n", prefix, author, msg.Text)
}

func (s *Session) handleInput(line string) {
	if line == "" {
		return
	}

	if strings.HasPrefix(line, "/") {
		s.handleCommand(line)
		return
	}

	_ = s.client.Send(OutgoingMessage{Action: "message", Text: line, Room: s.room})
}

func (s *Session) handleCommand(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/quit":
		close(s.stopChan)
		_ = s.client.Close()
		fmt.Println("Disconnected from chat.")
	case "/help":
		s.printHelp()
	case "/users":
		_ = s.client.Send(OutgoingMessage{Action: "users", Room: s.room})
		fmt.Println("Requested online users…")
	case "/pm":
		if len(parts) < 3 {
			fmt.Println("Usage: /pm <user> <message>")
			return
		}
		target := parts[1]
		text := strings.TrimPrefix(line, "/pm "+target+" ")
		_ = s.client.Send(OutgoingMessage{Action: "pm", To: target, Text: text, Room: s.room})
		fmt.Println("Private message sent.")
	case "/manga":
		if len(parts) < 2 {
			fmt.Println("Usage: /manga <manga-id>")
			return
		}
		newRoom := parts[1]
		s.switchRoom(newRoom)
	case "/history":
		s.printHistory()
	case "/status":
		_ = s.client.Send(OutgoingMessage{Action: "status", Room: s.room})
		fmt.Println("Requested room status…")
	default:
		fmt.Println("Unknown command. Type /help for assistance.")
	}
}

func (s *Session) switchRoom(room string) {
	close(s.stopChan)
	_ = s.client.Close()
	fmt.Printf("Switching to #%s…\n", room)
	s.room = room
	s.stopChan = make(chan struct{})
	s.client = NewWSClient(room)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.client.Connect(ctx); err != nil {
		fmt.Printf("✗ Unable to join #%s: %v\n", room, err)
		return
	}
	fmt.Printf("Connected to #%s. Type /help for commands.\n", room)

	go s.listen()
	go s.listenSignals()
}

func (s *Session) printHistory() {
	client := NewHistoryClient(s.verbose)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	history, raw, err := client.Fetch(ctx, s.room, 20)
	if s.verbose && len(raw) > 0 {
		fmt.Println(string(raw))
	}
	if err != nil {
		fmt.Println("✗ Unable to load chat history.")
		fmt.Println("Check if chat server is running: mangahub server status")
		return
	}

	RenderHistory(history, s.room, 20, s.quiet)
}

func (s *Session) printWelcome() {
	roomLabel := "#general"
	if s.room != "" {
		roomLabel = "#" + s.room
	}
	if !s.quiet {
		fmt.Printf("Connected to %s. Type /help for commands.\n", roomLabel)
	}
}

func (s *Session) printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("/help     Show this help message")
	fmt.Println("/users    List online users")
	fmt.Println("/pm USER  Send a private message")
	fmt.Println("/manga ID Switch to a manga-specific room")
	fmt.Println("/history  Show the last 20 messages")
	fmt.Println("/status   Show room status")
	fmt.Println("/quit     Leave chat")
}
