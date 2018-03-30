package pan

import (
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"time"
	"strings"
	"encoding/json"
	"fmt"
)

// read data from consul kv and watch it

type CenterConfig struct {
	Address 			string
	Scheme 				string
	Namespace 			string
	NamespaceMap 	 	string   //string map for generate the key path. like "/Configs/{namespace}/{key}"
	ContentType 		string   //type of consul value,  default json

	Key 				string
	keyIndex 			uint64
	watch 				bool   //default watch  true
	Interval 			time.Duration

}

func ConsulKV(config CenterConfig) *consulapi.KV {
	cconf := consulapi.DefaultConfig()
	cconf.Address = config.Address
	cconf.Scheme = config.Scheme
	client, err := consulapi.NewClient(cconf)
	if err != nil{
		log.Panicf("Error creating consul client", err)
	}
	kv := client.KV()
	return kv
}


func KVGet(key string, kv *consulapi.KV)([]byte,uint64,error){
	pair,meta,err := kv.Get(key,nil)
	if err != nil {
		log.Fatal(err)
	}
	if pair == nil {
		fmt.Println("Frist GeT Nil from center,Your key is:", key)
		return nil,uint64(0),nil
	}
	return pair.Value,meta.LastIndex,err
}

func WatchKey(key string, ch chan []byte, kv *consulapi.KV, keyIndex uint64) {
	currentIndex := keyIndex
	for {

		pair, meta, err := kv.Get(key, &consulapi.QueryOptions{
			WaitIndex: currentIndex,
		})
		if err != nil {
			fmt.Println("Error for get key,I will sleep 2 mins:",err)
			time.Sleep(2 * time.Minute)

		}

		if pair == nil || meta == nil {
			// Query won't be blocked if key not found
			//time.Sleep(1 * time.Second)
		} else {
			ch <- pair.Value
			currentIndex = meta.LastIndex
		}

	}
}
func WatchKeyWithInterval(key string, ch chan []byte, kv *consulapi.KV, keyIndex uint64, interval time.Duration) {
	currentIndex := keyIndex
	for {
		pair, meta, err := kv.Get(key,nil)
		if err != nil {
			fmt.Println("Error for get key,I will sleep another 2 mins:",err)
			time.Sleep(2 * time.Minute)
		}

		if pair == nil || meta == nil {
			// Query won't be blocked if key not found
			//time.Sleep(1 * time.Second)
		} else if meta.LastIndex != currentIndex {
			ch <- pair.Value
			currentIndex = meta.LastIndex
		}
		time.Sleep(interval)

	}
}

func (p *Pan) ReadCenterWithWatch()  {

	kv := ConsulKV(p.CenterConfig)
	ch := make(chan []byte)
	key := Sformat(p.CenterConfig.NamespaceMap,"{namespace}",p.CenterConfig.Namespace, "{key}",p.CenterConfig.Key)
	centerMap := make(map[string]interface{})

	//first get key from center
	data,i1,err := KVGet(key,kv)
	if err != nil{
		fmt.Println("error when get key :",err)
	}else {
		json.Unmarshal(data, &centerMap)
		p.center = UpMapKey(&centerMap)
		// watch

		go WatchKeyWithInterval(key, ch, kv, i1, p.CenterConfig.Interval)
		go func() {
			for data := range ch {
				json.Unmarshal(data, &centerMap)
				p.center = UpMapKey(&centerMap)
			}
		}()
	}

}

func Sformat(str string, args ...string) string {
	r := strings.NewReplacer(args...)
	ss := r.Replace(str)
	return ss
}
