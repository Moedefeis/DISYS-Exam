package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	proto "github.com/Moedefeis/DISYS-Exam/grpc"
	"google.golang.org/grpc"
)

var (
	ownPort     int
	leaderPort  int = 9999
	replicaPort int = 8888
	replica     proto.DictionaryClient
	dictionary  = make(map[string]string)
	lock        = make(chan bool)
)

type Dictionary struct {
	proto.UnimplementedDictionaryServer
}

func (d *Dictionary) Add(ctx context.Context, request *proto.AddRequest) (*proto.AddResponse, error) {
	<-lock
	success := true
	dictionary[request.Word] = request.Def

	// If the node is the leader, call add on the replica node
	if isLeader() {
		response, err := replica.Add(ctx, request)
		if err != nil || !response.Success {
			success = false
		}
	}
	//log.Println("Added: " + request.Word + " - " + request.Def)
	lock <- true
	return &proto.AddResponse{Success: success}, nil
}

func (d *Dictionary) Read(ctx context.Context, request *proto.ReadRequest) (*proto.ReadResponse, error) {
	<-lock
	def := dictionary[request.Word]
	log.Printf("Read " + request.Word + ", returned " + def)
	lock <- true
	return &proto.ReadResponse{Def: def}, nil
}

func (d *Dictionary) Crashed(ctx context.Context, serverID *proto.ServerID) (*proto.Void, error) {
	<-lock

	// If the crashed node was the leader, elect itself as new leader. This only works for two nodes,
	// since if the leader is crashed, the single replica node can elect itself as the leader.

	if int(serverID.Id) == leaderPort {
		leaderPort = ownPort
	}
	lock <- true
	return &proto.Void{}, nil
}

func isLeader() bool {
	return ownPort == leaderPort
}

func main() {
	ownPort, _ = strconv.Atoi(os.Args[1])

	if isLeader() {
		conn, err := grpc.Dial(fmt.Sprintf(":%v", replicaPort), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
		}
		replica = proto.NewDictionaryClient(conn)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))

	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", ownPort, err)
	}

	defer listener.Close()

	server := grpc.NewServer()
	proto.RegisterDictionaryServer(server, &Dictionary{})

	log.Println("Server open")

	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}
