package pan

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"strings"
	"os"
	"github.com/spf13/cast"
	"time"
)

// thank you for viper by spf13
// Pan take

// config   from file yaml or json file
// env  from os.env
// flag    command line args
// center   from remote kv store like consul
// overwrite  write by hand   use for test or debug

type Pan struct {

	// Delimiter that separates a list of keys
	// used to access a nested value in one go
	keyDelim string


	configFile 		string //config path with file name
	configType 		string //config type json or yaml
	envPrefix 		string

	CenterConfig    CenterConfig

	config 			map[string]interface{}
	env 			map[string]interface{}
	flag 			map[string]interface{} //Todo flag
	center 			map[string]interface{}
	override 		map[string]interface{}

	typeByDefValue  bool
}


var p *Pan

var supportTypes = []string{"json","yaml"}
func init()  {
	p = new(Pan)
}

func New() *Pan  {
	p := new(Pan)
	p.keyDelim = "."
	p.config = make(map[string]interface{})
	p.env = make(map[string]interface{})
	p.center = make(map[string]interface{})
	p.override = make(map[string]interface{})
	p.typeByDefValue = true
	// set for Eigen
	//p.CenterConfig.Address = "consul.aidigger.com:8500"
	p.CenterConfig.Address = "consul.aidigger.com:8500"
	p.CenterConfig.Scheme = "http"
	p.CenterConfig.Namespace = "debug"
	p.CenterConfig.NamespaceMap = "/Configs/{namespace}/{key}"
	p.CenterConfig.Interval = 60 * time.Second

	return p
}

// set config file

func (p *Pan) SetConfigFile(path string, fileType string)  {
	if path != "" {
		p.configFile = path
	}
	for _, b := range supportTypes{
		if b == fileType {
			p.configType = fileType
		}
	}
	if p.configType == ""{
		panic("config type not supported,please use: json or yaml")
	}
}

func (p *Pan)ReadInConfig() error {
	conf := make(map[string]interface{})
	data,err := ioutil.ReadFile(p.configFile)

	if err != nil {
		return err
	}

	switch p.configType {
	case "json":
		json.Unmarshal(data,&conf)
	case "yaml":
		//TODO support yaml
		fmt.Println("jok")
	}
	for k,_ := range conf {
		k = strings.ToUpper(k)
	}
	p.config = UpMapKey(&conf)
	return nil
}

func (p *Pan) Get(key string) interface{}  {
	key = strings.ToUpper(key)
	res := p.find(key)

	// default json num type is float64
	if p.typeByDefValue{
		switch res.(type) {
		case string:
			return cast.ToString(res)
		case int64:
			return cast.ToInt64(res)
		case map[string]interface{}:
			return cast.ToStringMap(res)
		case int32:
			return cast.ToInt32(res)
		case int8:
			return cast.ToInt8(res)
		case int:
			return cast.ToInt(res)
		case float64:
			return cast.ToFloat64(res)
		case float32:
			return cast.ToFloat32(res)
		case bool:
			return cast.ToBool(res)
		case []string:
			return cast.ToStringSlice(res)

		}
	}
	return res
}

// get string value
func (p *Pan) GetStr(key string) string  {
	return cast.ToString(p.Get(key))
}

// get string slice

func (p *Pan) GetStrSlice(key string)[]string  {
	return cast.ToStringSlice(p.Get(key))
}
// get int
func (p Pan) GetInt(key string) int  {
	return cast.ToInt(p.Get(key))
}

//get bool
func (p Pan) GetBool(key string) int  {
	return cast.ToBool(p.Get(key))
}

// env

func (p *Pan)SetEnvPrefix(prefix string)  {
	if prefix != ""{
		p.envPrefix = prefix
	}
}

func (p *Pan) ReadAllEnv()  {
	envs := make(map[string]interface{})
	sysEnvlist := os.Environ()
	for _,e := range sysEnvlist{
		tempList := strings.Split(e,"=")
		envs[strings.ToUpper(tempList[0])] = tempList[1]
	}
	p.env = envs
}

func (p *Pan) ReadEnvWithPrefix()  {
	//Todo readEnvWithPrefix
}

func UpMapKey(m *map[string]interface{}) map[string]interface{}  {

	tempMap := make(map[string]interface{})
	for k,v := range *m{
		tempMap[strings.ToUpper(k)] = v
	}

	return tempMap
}
// search map

//todo
func (p *Pan) searchMap(source map[string]interface{}, path []string) interface{}  {

	if len(path) == 0 {
		return source
	}
	next,ok := source[path[0]]

	if ok {
		// fast path
		if len(path) == 1 {
			return next
		}
		
		//nested case

		switch next.(type) {
		case map[string]interface{}:
			newNext := cast.ToStringMap(next)
			return p.searchMap(UpMapKey(&newNext), path[1:])
		default:
			fmt.Println(next)
			return nil
		}
	}
	
	return nil
}


// Given a key, find the value.
// Pan will check in the following order:
// overwrite center env config
// Note: this assumes a up-cased key given.

func (p *Pan) find(key string) interface{}  {

	var(
		val interface{}
		path = strings.Split(key, p.keyDelim)
		//nested = len(path) > 1
		)
	// override

	val = p.searchMap(p.override,path)
	if val != nil{
		return val
	}
	//env
	val = p.searchMap(p.env,path)
	if val != nil{
		return val
	}
	//center

	val = p.searchMap(p.center,path)
	if val != nil{
		return val
	}

	// config
	val = p.searchMap(p.config,path)
	if val != nil{
		return val
	}

	return nil
}