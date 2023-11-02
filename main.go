package main

import (
	"context"
	"log"
	"bytes"
	"strings"
	"net"
	"net/http"
	"net/http/httptest"
	//"google.golang.org/protobuf/reflect/protoreflect"
	"fmt"
	// "net/url"
	"os"
	"io/ioutil"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/aidenwallis/go-write/write"
	// "github.com/golang/protobuf/ptypes"
 

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/felixge/httpsnoop"

	servicepb "github.com/liuyuxiao/cache_server/proto/cache_server_service"

	"github.com/liuyuxiao/cache_server/funcs"
)
var Global = "myvalue" // Go全局变量
type svrConfig struct {
    IP				string `json:"ip"`
    RpcPort			string `json:"port_rpc"`
    GatePort		string `json:"port_gate"`
	RpcServerList 	string `json:"all_grpc"`
}

type server struct{
	servicepb.UnimplementedCacheServerServer
}


var kvDict = make(map[string]string)
var rpcServerList []string
func NewServer() *server {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *servicepb.HelloRequest) (*servicepb.HelloReply, error) {
	Global = "a:a"
	log.Println("hello")
	return &servicepb.HelloReply{Message: in.Name + Global}, nil
}

func (s *server) AddKeyValue(ctx context.Context, in *servicepb.AddKeyValueRequest) (*servicepb.AddKeyValueReply, error) {
	kvDict[in.Key] = in.Value
	fmt.Println("log ... succeed in writing " + kvDict[in.Key] + " into " + in.Key)
	return &servicepb.AddKeyValueReply{Message: "succeed in writing " + kvDict[in.Key] + " into " + in.Key}, nil
}


func (s *server) GetValueByKey(ctx context.Context, in *servicepb.GetValueRequest) (*servicepb.GetValueReply, error) {
	v, ok := kvDict[in.Key]
	log.Println("log ... get ", v," from ", in.Key, " with interControl:", in.InterControl)
	isExist := "false"
	if ok {
		isExist = "true"
	}else{
		if in.InterControl == 0{
			for i := 0; i < len(rpcServerList); i++ {
				addr := rpcServerList[i]
				log.Println(addr)
				var result = make(map[string]string)
				funcs.CallGrpcSever(&result, addr, in.Key, "GetValueByKey")
				log.Println("out:::", result["IsExist"])
				if result["IsExist"] == "true"{
					v = result["Value"]
					isExist = "true"
					break
				}
			}
		}
	}
	return &servicepb.GetValueReply{Key: in.Key, Value:  v, IsExist: isExist} , nil
}

func (s *server) DeleteValueByKey(ctx context.Context, in *servicepb.DeleteValueRequest) (*servicepb.DeleteValueReply, error) {
	_, ok := kvDict[in.Key]
	// log.Println("log ... delete ", v," from ", in.Key, " with interControl:", in.InterControl)
	num := 0
	if ok {
		delete(kvDict, in.Key)
		num = 1
	}else{
		if in.InterControl == 0{
			// for i := 0; i < len(rpcServerList); i++ {
			// 	addr := rpcServerList[i]
			// 	log.Println(addr)
			// 	var result = make(map[string]string)
			// 	funcs.CallGrpcSever(&result, addr, in.Key, "GetValueByKey")
			// 	log.Println("out:::", result["IsExist"])
			// 	if result["IsExist"] == "true"{
			// 		v = result["Value"]
			// 		isExist = "true"
			// 		break
			// 	}
			// }
		}
	}
	return &servicepb.DeleteValueReply{Num: int64(num)} , nil
}


func setProxyRules(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" && request.URL.Path == "/"{
			bodyBytes, _ := ioutil.ReadAll(request.Body)
			var payload map[string]interface{}
			json.Unmarshal(bodyBytes, &payload)
			var lastKeyInJson string
			for k :=range payload{
				lastKeyInJson = k
			}
			log.Println("valueType:",funcs.CheckType(payload[lastKeyInJson]))
			payload["value"] = funcs.EncodeToString(payload[lastKeyInJson])
			payload["key"] = lastKeyInJson
			bodyBytes, _ = json.Marshal(payload)
			request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			m:=httpsnoop.CaptureMetrics(handler,writer,request) //request must be used
			log.Printf("http[%d]-- %s -- %s\n",m.Code,m.Duration,request.URL.Path)
		}
		if request.Method == "GET" && funcs.MatchURLPath(request.URL.Path, "/*"){
			writerFake := httptest.NewRecorder()
			// m:=httpsnoop.CaptureMetrics(handler,writerFake,request) //use a fake response to get the result
			// log.Printf("http[%d]-- %s -- %s\n",m.Code,m.Duration,request.URL.Path)
			rww := funcs.NewResponseWriterWrapper(writerFake)
			handler.ServeHTTP(rww, request) //copy the result from the fake response to buf; but it sends an empty request to the grpc server
			writer.Header()
			var msgJson map[string]string
			json.Unmarshal([]byte(rww.Body.String()), &msgJson)
			fmt.Println("GET KEY:",msgJson["key"])
			fmt.Println("IF  GET:",msgJson["isExist"])
			fmt.Println("GET VAL:",msgJson["value"])
			
			if msgJson["isExist"] == "false"{
				writer.WriteHeader(404)
			}else{
				valueFromDecoding :=  funcs.DecodeFromString(msgJson["value"])
				typeofValue := funcs.CheckType(valueFromDecoding)
				writer.WriteHeader(200)
				switch typeofValue{
					case "[]interface {}":
						resJson := make(map[string][]interface{})
						resJson[msgJson["key"]] = funcs.DecodeFromString(msgJson["value"]).([]interface{})
						json.Marshal(resJson)
						write.New(writer, http.StatusTeapot).JSON(&resJson)
					case "int":
						resJson := make(map[string]int64)
						resJson[msgJson["key"]] = funcs.DecodeFromString(msgJson["value"]).(int64)
						json.Marshal(resJson)
						write.New(writer, http.StatusTeapot).JSON(&resJson)
					case "bool":
						resJson := make(map[string]bool)
						resJson[msgJson["key"]] = funcs.DecodeFromString(msgJson["value"]).(bool)
						json.Marshal(resJson)
						write.New(writer, http.StatusTeapot).JSON(&resJson)
					case "double":
						resJson := make(map[string]float64)
						resJson[msgJson["key"]] = funcs.DecodeFromString(msgJson["value"]).(float64)
						json.Marshal(resJson)
						write.New(writer, http.StatusTeapot).JSON(&resJson)
					case "string":
						resJson := make(map[string]string)
						resJson[msgJson["key"]] = funcs.DecodeFromString(msgJson["value"]).(string)
						json.Marshal(resJson)
						write.New(writer, http.StatusTeapot).JSON(&resJson)
				}
				
		}
			// write.New(writer, http.StatusTeapot).Empty()
			//write the result extracted from buf into the real response
		}
		if request.Method == "DELETE" && funcs.MatchURLPath(request.URL.Path, "/*"){
			m:=httpsnoop.CaptureMetrics(handler,writer,request) //use a fake response to get the result
			log.Printf("http[%d]-- %s -- %s\n",m.Code,m.Duration,request.URL.Path)
		}

    })
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

	rpcServerListTmp := strings.Split(cfg.RpcServerList,",")
	for i:=0;i<len(rpcServerListTmp);i++{
		if rpcServerListTmp[i] != cfg.IP + ":" + cfg.RpcPort{
		// log.Println(rpcServerListTmp[i], cfg.IP + ":" + cfg.RpcPort)
		rpcServerList = append(rpcServerList, rpcServerListTmp[i])
		}
	}

	lis, err := net.Listen("tcp", ":" + cfg.RpcPort)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the CacheServer service to the server
	servicepb.RegisterCacheServerServer(s, &server{})
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

	gwmux := runtime.NewServeMux(
		// runtime.WithForwardResponseOption(func (ctx context.Context, writer http.ResponseWriter, resp protoreflect.ProtoMessage) error {
		// 	writer.Write([]byte("；你好"))
		// 	log.Println("withforwardresponse")
		// 	return nil
		// }),
	)
	// Register CacheServer
	err = servicepb.RegisterCacheServerHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":" + cfg.GatePort,
		// Handler: gwmux,
		Handler: setProxyRules(gwmux),
	}

	log.Println("Serving gRPC-Gateway on " + cfg.IP + ":" + cfg.GatePort)
	log.Fatalln(gwServer.ListenAndServe())
}
