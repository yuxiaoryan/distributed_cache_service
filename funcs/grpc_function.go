package funcs
import (
	"context"
	"fmt"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	servicepb "github.com/liuyuxiao/cache_server/proto/cache_server_service"
)

func CallGrpcSever(result *map[string]string, addr string, key string, funcName string) {
	
	conn, err := grpc.Dial(addr , grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("connect error to: ", addr)
	}
	defer conn.Close()
	c := servicepb.NewCacheServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if funcName == "GetValueByKey"{
		(*result)["Value"] = ""
		(*result)["IsExist"] = "false"
		r, err := c.GetValueByKey(ctx, &servicepb.GetValueRequest{Key: key, InterControl: 1})
		
		if err != nil {
			fmt.Println("can not connect to: ",addr)
		} else {
			if r.IsExist == "true"{
				(*result)["Value"] = r.Value
				(*result)["IsExist"] = "true"
			}
		}
	}	
	if funcName == "DeleteValueByKey"{
		(*result)["num"] = ""
		r, err := c.DeleteValueByKey(ctx, &servicepb.DeleteValueRequest{Key: key, InterControl: 1})
		if err != nil {
			fmt.Println("can not connect to: ",addr)
		} else {
			(*result)["num"] = Any2String(r.Num)
		}
	}	
}