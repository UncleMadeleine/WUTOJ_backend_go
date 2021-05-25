package helper

import (
	"github.com/gin-gonic/gin"
	"reflect"
	"strings"
)

type ReturnType struct {
	Status int
	Msg    string
	Data   interface{}
}

// 获取语言类型
func LanguageID(language string) int {
	var id int
	switch language {
	case "c.gcc":
		id = 0
	case "cpp.g++":
		id = 1
	case "java.openjdk8":
		id = 2
	case "python.cpython3.6":
		id = 3
	default:
		id = -1
	}
	return id
}

// 获取语言类型
func LanguageType(typeInt int) string {
	var language string
	switch typeInt {
	case 0:
		language = "c.gcc"
	case 1:
		language = "cpp.g++"
	case 2:
		language = "java.openjdk8"
	case 3:
		language = "python.cpython3.6"
	default:
		language = "undefined"
	}
	return language
}

// 结构体转换为map
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		mapValue := v.Field(i).Interface()
		// 递归获取数据
		if reflect.TypeOf(mapValue).Kind() == reflect.Struct {
			if v.Field(i).Type().String() != "time.Time" {
				innerMap := Struct2Map(mapValue)
				for key, value := range innerMap {
					data[key] = value
				}
			}
		}
		// 转换驼峰为下划线
		upperField := t.Field(i).Name
		field := ""
		index := 0
		for j := 0; j < len(upperField)-1; j++ {
			if (upperField[j] >= 'a' && upperField[j] <= 'z') &&
				(upperField[j+1] >= 'A' && upperField[j+1] <= 'Z') {
				field += upperField[index:j+1] + "_"
				index = j + 1
			}
		}
		field += upperField[index:]
		data[strings.ToLower(field)] = v.Field(i).Interface()
	}
	return data
}

// 模块内统一返回格式
func ReturnRes(status int, msg string, data interface{}) ReturnType {
	returnType := ReturnType{status, msg, data}
	return returnType
}

func ApiReturn(status int, msg string, data interface{}) gin.H {
	return gin.H{
		"status":  status,
		"message": msg,
		"data":    data,
	}
}

func BackendApiReturn(status int, msg string, data interface{}) gin.H {
	return gin.H{
		"status": status,
		"msg":    msg,
		"data":   data,
	}
}
