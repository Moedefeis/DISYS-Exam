package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	proto "github.com/Moedefeis/DISYS-Exam/grpc"
	"google.golang.org/grpc"
)

var (
	conns  = make(map[int]proto.DictionaryClient)
	ctx    = context.Background()
	reader *bufio.Scanner
)

func connect(port int) {
	conn, err := grpc.Dial(fmt.Sprintf("%v", port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}
	defer conn.Close()

	client := proto.NewDictionaryClient(conn)
	conns[port] = client
}

func add(word string, definition string) {
	request := proto.AddRequest{Word: word, Def: definition}
	//log.Println("Adding: " + word + " - " + definition)

	// Call add on both nodes. The leader node will update the replica.

	for port, conn := range conns {
		response, err := conn.Add(ctx, &request)
		if err != nil || !response.Success {
			conn.Crashed(ctx, &proto.ServerID{Id: int32(port)})
		}
	}
}

func read(word string) string {
	request := proto.ReadRequest{Word: word}
	var def string

	// Call read on both nodes

	for port, conn := range conns {
		response, err := conn.Read(ctx, &request)
		if err != nil {
			conn.Crashed(ctx, &proto.ServerID{Id: int32(port)})
		} else {
			def = response.Def
		}
	}

	return def
}

func handleInput() {
	for {
		reader.Scan()
		input := reader.Text()

		fields := strings.Fields(input)
		op := strings.ToLower(fields[0])

		if op == "add" {
			add(fields[1], fields[2])
		} else if op == "read" {
			def := read(fields[1])
			log.Printf(def)
		} else {
			log.Printf("Invaild input")
		}
	}
}

func main() {
	reader = bufio.NewScanner(os.Stdin)

	// Connect to both nodes
	connect(9999)
	connect(8888)

	add("hej", "med dig")
	log.Printf(read("hej"))

	//handleInput()
}
