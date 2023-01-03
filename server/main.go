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
	ctx         = context.Background()
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
	if isLeader() {
		response, err := replica.Add(ctx, request)
		if err != nil || !response.Success {
			success = false
		}
	}
	lock <- true
	return &proto.AddResponse{Success: success}, nil
}

func (d *Dictionary) Read(ctx context.Context, request *proto.ReadRequest) (*proto.ReadResponse, error) {
	<-lock
	def := dictionary[request.Word]
	lock <- true
	return &proto.ReadResponse{Def: def}, nil
}

func (d *Dictionary) Crashed(ctx context.Context, serverID *proto.ServerID) (*proto.Void, error) {
	<-lock
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
}
