package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DataSet struct {
	energy []float64
	sig []float64
}

func main() {
	url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9022734&req=2566&PenSectID=13660487&json"
	bytesBody := request(&url)
	var results map[string]interface{}
	err := json.Unmarshal(bytesBody, &results) // parsing json
	fmt.Println(err)
	pts := results["datasets"].([]interface{})[0].(map[string]interface{})["pts"].([]interface{}) // accessing pts map as a slice
	fmt.Printf("%T\n", pts)
	ds := &DataSet{}
	fullfilingStruct(pts, ds)
	fmt.Println(*ds)
}

func fullfilingStruct(intr []interface{}, DS *DataSet) {
	// intrPrefix := intr.(map[string]interface{})
	for _, v := range intr {
		// fmt.Println(v.(map[string]interface{})["E"])
		DS.energy = append(DS.energy, v.(map[string]interface{})["E"].(float64))
		DS.sig = append(DS.sig, v.(map[string]interface{})["Sig"].(float64))
	}
}

func request(url *string) []byte {
	get, err := http.Get(*url)
	if err != nil {
		panic(err)
	}
	defer get.Body.Close()
	body := get.Body
	bytesBody, _ := ioutil.ReadAll(body)
	return bytesBody
}