package pan

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
)

// thank you for viper by spf13
// Pan take

// config   from file yaml or json file
// env  from os.env
// flag    command line args
// center   from remote kv store like consul
// overwrite  write by hand   use for test or debug

type Pan struct {

	configFile 		string //config path with file name
	configType 		string //config type json or yaml

	config map[string]interface{}

}

var p *Pan

var supportTypes = []string{"json","yaml"}
func init()  {
	p = new(Pan)
}

func New() *Pan  {
	p := new(Pan)
	p.config = make(map[string]interface{})
	return p
}


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
		//TODO
		fmt.Println("jok")
	}
	p.config = conf
	return nil
}

func (p *Pan) Get(key string) interface{}  {
	return p.config[key]
}