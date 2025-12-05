package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"time"
)

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

type Message struct {
	User string
	Text string
	Time time.Time
}

type UpdateReply struct {
	Messages []Message
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	nameRaw, _ := reader.ReadString('\n')
	name := strings.TrimSpace(nameRaw)
	if name == "" {
		name = "anonymous"
	}

	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var subReply SubscribeReply
	err = client.Call("Chat.Subscribe", SubscribeArgs{User: name}, &subReply)
	if err != nil {
		panic(err)
	}

	clientID := subReply.ClientID
	fmt.Println("Connected with client ID:", clientID)

	go func() {
		for {
			var upd UpdateReply
			client.Call("Chat.GetUpdates", UpdateArgs{ClientID: clientID}, &upd)

			for _, m := range upd.Messages {
				fmt.Printf("[%s] %s: %s\n",
					m.Time.Format("15:04:05"), m.User, m.Text)
			}

			time.Sleep(300 * time.Millisecond)
		}
	}()

	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		text := strings.TrimSpace(line)

		if text == "exit" {
			fmt.Println("Goodbye!")
			return
		}

		args := SendArgs{User: name, Text: text}
		client.Call("Chat.SendMessage", args, &struct{}{})
	}
}
