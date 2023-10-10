package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"io/ioutil"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	servicepb "github.com/liuyuxiao/cache_server/proto/cache_server_service"
)

type svrConfig struct {
    IP			string `json:"ip"`
    RpcPort		string `json:"port_rpc"`
    GatePort	string `json:"port_gate"`
}

type server struct{
	servicepb.UnimplementedGreeterServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *servicepb.HelloRequest) (*servicepb.HelloReply, error) {
	return &servicepb.HelloReply{Message: in.Name + " world11"}, nil
}

func (s *server) AddKeyValue(ctx context.Context, in *servicepb.AddKeyValueRequest) (*servicepb.AddKeyValueReply, error) {
	return &servicepb.AddKeyValueReply{Message: in.Key+ ": " + in.Value}, nil
}

func main() {

	// Load ip and ports from config.json
	cfg := svrConfig{}
	orgData, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Failed to config:", err)
	}
	data := os.ExpandEnv(string(orgData))

	err = json.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Fatalln("Failed to load cfg:", err)
	}

	lis, err := net.Listen("tcp", ":" + cfg.RpcPort)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	servicepb.RegisterGreeterServer(s, &server{})
	// Serve gRPC server
	log.Println("Serving gRPC on " + cfg.IP + ":" + cfg.RpcPort)
	go func() {
		log.Fatalln(s.Serve(lis))
	}()
	
	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		cfg.IP + ":" + cfg.RpcPort,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = servicepb.RegisterGreeterHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + cfg.GatePort,
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on " + cfg.IP + ":" + cfg.GatePort)
	log.Fatalln(gwServer.ListenAndServe())
}

//To do: 改hello_world这一愚蠢的项目名然后更新源文件；go的字典 