## Pan

## the simple config managerment library

## feature

* file //json or yaml
* center //consul



## priority

* overwrite 
* env 
* center 
* config


## demo 

```go
package main

import "github.com/GoSome/pan"


func InitConfig() *pan.Pan {
	p := pan.New()
	p.SetConfigFile("./config.json","json")
	p.ReadInConfig()
	p.CenterConfig.Namespace = "test"
	p.CenterConfig.Address = "192.168.3.14:8500"
	p.CenterConfig.Key = "config.json"
	p.ReadCenterWithWatch()
	return p
}



func main()  {
    var Config = InitConfig()
    Config.GetStrSlice("key1")
    Config.GetStr("key2")
}
```
