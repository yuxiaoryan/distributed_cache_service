package funcs

import (
	"fmt"
	"strconv"
	"log"
	"strings"
)
const resevredStr = "#!*+^%"
const resevredStrLen = len(resevredStr) + 1
func Test(){
	fmt.Println("test")
}
func MatchURLPath(path string, mode string) bool{
	pathSlice := strings.Split(path, "/")
	modeSlice := strings.Split(mode, "/")
	if len(pathSlice) != len(modeSlice){
		return false
	}
	for i:=0;i<len(pathSlice) - 1;i++{
		if pathSlice[i] != modeSlice[i]{
			return false
		}
	}
	return true
}
func IsConsideredDateType(a byte) bool{
	if a=='i' || a=='f' ||a=='d' ||a=='s' ||a=='b'{
		return true
	}
	return false
}
func CheckType(v interface{}) string {
	s := fmt.Sprintf("%T", v)
	if s == "float64"{
		s = "double"
	}
	if s == "float32"{
		s = "float"
	} 
	if s == "int64"{
		s = "int"
	}
    return s
}
func Any2String(arg interface{}) string{
	switch arg.(type) {
		case int:
			return fmt.Sprintf("%d", arg) 
		case int64:
			return fmt.Sprintf("%d", arg) 
		case string:
			return arg.(string)
		case float64:
			return fmt.Sprintf("%.5f" ,arg.(float64))
		case bool:
			if arg.(bool){
				return "true"
			}else{
				return "false"}
		default:
			return ""
	}
}
func EncodeToString(v interface{}) string {
	typeofValue := CheckType(v)
	res := ""
	dataType := []byte("s" + resevredStr)
	// fmt.Println(typeofValue)
	switch typeofValue {
		case "bool":
			dataType[0] = 'b'
			res = string(dataType) + Any2String(v)
		case "string":
			dataType[0] = 's'
			res = string(dataType) + Any2String(v)
		case "int":
			dataType[0] = 'i'
			res = string(dataType) + Any2String(v)
		case "double":
			dataType[0] = 'd'
			res = string(dataType) + Any2String(v)
		case "float":
			dataType[0] = 'f'
			res = string(dataType) + Any2String(v)
		case "[]interface {}":
			a := v.([]interface {})
			res = "["
			for i:=0;i<len(a);i++{
				dataType[0] = CheckType(a[i])[0]
				comma := ","
				if i == len(a) - 1{
					comma = ""
				}
				var content string
				if IsConsideredDateType(dataType[0]){
					content = Any2String(a[i])
					res = res + string(dataType) + content + comma
				}else {
					content =""
					res = res + "s" + resevredStr + content + comma
				}
				// res = res + string(dataType) + content + comma
				// switch dataType[0] {  
				// 	case 'b':
				// 		res = res + string(dataType) + content + comma
				// 	case 'i':
				// 		res = res + string(dataType) + content + comma
				// 	case 's':
				// 		res = res + string(dataType) + content + comma
				// 	case 'd':
				// 		res = res + string(dataType) + content + comma
				// 	default:
				// 		res = res + "s" + resevredStr + comma
				// }
			}
			res = res + "]"
	}
    return res
}

func matchReservedStr(str string) bool{
	if IsConsideredDateType(str[0]) && str[1:] == resevredStr{
		return true
	}
	return false
}
func findEndComma(runeStr []rune, left int, right int) int {
	end := right + 1
	for i:=left; i <= right; i++{
		if string(runeStr[i]) == ","{
			if i == right{
				end = i
				break
			}
			//fmt.Println(string(runeStr[i + 1:i + 1 + resevredStrLen]))
			if matchReservedStr(string(runeStr[i + 1:i + 1 + resevredStrLen])){
				end = i
				break
			}
		}
	}
	return end
	
}


func String2Any(dataType byte ,content string) interface{}{
	var item interface{}
	switch dataType{
		case 'i':
			distInt64, err :=strconv.ParseInt(content, 10, 64)
			if err != nil{
				log.Fatal("Error in trans from str to int:", err)
			}
			item = distInt64
		case 's':
			item = content
		case 'd':
			distFloat64, err :=strconv.ParseFloat(content, 64)
			if err != nil{
				log.Fatal("Error in trans from str to float64:", err)
			}
			item = distFloat64
			
		case 'f':
			distFloat32, err :=strconv.ParseFloat(content, 32)
			if err != nil{
				log.Fatal("Error in trans from str to float64:", err)
			}
			item = distFloat32
		case 'b':
			if content == "true"{
				item = true
			}else{
				item = false
			}
	}
	return item
}

func DecodeFromString(str string) interface{} {
	var res interface{}

	if str[0] == '['{
		arr := make([]interface{},0)
		var item interface{}
		i := 1
		runeStr := []rune(str)
		for true{
			if i > len(runeStr) - 2{ //len(str)-1 is the index of the last char in the str except ']' 
				break
			}
			dataType := string(runeStr[i])[0]
			i += resevredStrLen
			endCommaIndex := findEndComma(runeStr, i, len(runeStr) - 2)
			content := string(runeStr[i:endCommaIndex])
			if i == endCommaIndex{//A special check for the situation where the last item is empty
				content = ""
			}
			//fmt.Println(content)
			item = String2Any(dataType, content)
			arr = append(arr,item) 
			i = endCommaIndex + 1
			if i >= len(runeStr) - 1{
				break
			} 
		}
		res = arr
	}else{
		if IsConsideredDateType(str[0]){
			res = String2Any(str[0], string([]rune(str)[resevredStrLen:]))
		}else{
			res = ""
		}
	}
	return res
}


// func MyPrintf(key string, args ...interface{}) {
// 	switch args[0].(type) {
// 		case int:
// 			fmt.Println(args[0], "is int")
// 		case string:
// 			fmt.Println(args[0], "is string")
// 		case float64:
// 			fmt.Println(args[0], "is float64")
// 		case bool:
// 			fmt.Println(args[0], " is bool")
// 		default:
// 			fmt.Println("未知的类型")
// 	}
//     // for _, arg := range args { //迭代不定参数
//     //     switch arg.(type) {
//     //     case int:
//     //         fmt.Println(arg, "is int")
//     //     case string:
//     //         fmt.Println(arg, "is string")
//     //     case float64:
//     //         fmt.Println(arg, "is float64")
//     //     case bool:
//     //         fmt.Println(arg, " is bool")
//     //     default:
//     //         fmt.Println("未知的类型")
//     //     }
//     // }
// }

