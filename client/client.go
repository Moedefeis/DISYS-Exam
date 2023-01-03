package main

import (
	"context"
	"fmt"
	"log"

	proto "github.com/Moedefeis/DISYS-Exam/grpc"
	"google.golang.org/grpc"
)

var (
	conns = make(map[int]proto.DictionaryClient)
	ctx   = context.Background()
)

func connect(port int) proto.DictionaryClient {
	conn, err := grpc.Dial(fmt.Sprintf("%v", port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	}

	return proto.NewDictionaryClient(conn)
}

func add(word string, definition string) {
	request := proto.AddRequest{Word: word, Def: definition}

	for port, conn := range conns {
		response, err := conn.Add(ctx, &request)
		if !response.Success || err != nil {
			conn.Crashed(ctx, &proto.ServerID{Id: int32(port)})
		}
	}
}

func read(word string) string {
	request := proto.ReadRequest{Word: word}
	var def string

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

func main() {
	conns[9999] = connect(9999)
	conns[8888] = connect(8888)
}
