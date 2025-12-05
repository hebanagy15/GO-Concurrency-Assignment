package main

import (
	"errors"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Message struct {
	User string
	Text string
	Time time.Time
}

type SendArgs struct {
	User string
	Text string
}

type SubscribeArgs struct {
	User string
}

type SubscribeReply struct {
	ClientID int
}

type UpdateArgs struct {
	ClientID int
}

type UpdateReply struct {
	Messages []Message
}

type Chat struct {
	mu       sync.Mutex
	clients  map[int]chan Message 
	nextID   int
}

func NewChat() *Chat {
	return &Chat{
		clients: make(map[int]chan Message),
	}
}

// SendMessage -> broadcast to all EXCEPT sender
func (c *Chat) SendMessage(args SendArgs, _ *struct{}) error {
	if args.Text == "" {
		return errors.New("empty message")
	}

	msg := Message{
		User: args.User,
		Text: args.Text,
		Time: time.Now(),
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, ch := range c.clients {
		if args.User == "" || args.User != "" {
		
			if args.User == "" {
				continue
			}
		}
		ch <- msg
	}

	return nil
}

func (c *Chat) Subscribe(args SubscribeArgs, reply *SubscribeReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++

	ch := make(chan Message, 20)
	c.clients[id] = ch
	reply.ClientID = id

	joinMsg := Message{
		User: "SYSTEM",
		Text: "User " + args.User + " joined",
		Time: time.Now(),
	}

	for cid, ch := range c.clients {
		if cid != id {
			ch <- joinMsg
		}
	}

	return nil
}

func (c *Chat) GetUpdates(args UpdateArgs, reply *UpdateReply) error {
	c.mu.Lock()
	ch, ok := c.clients[args.ClientID]
	c.mu.Unlock()

	if !ok {
		return errors.New("invalid client ID")
	}

out:
	for {
		select {
		case msg := <-ch:
			reply.Messages = append(reply.Messages, msg)
		default:
			break out
		}
	}

	return nil
}

func main() {
	chat := NewChat()

	err := rpc.Register(chat)
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	println("Chat server listening on :1234")
	for {
		conn, err := l.Accept()
		if err != nil {
			println("accept error:", err.Error())
			continue
		}
		go rpc.ServeConn(conn)
	}
}
