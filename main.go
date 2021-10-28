package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"bufio"
	"strings"
)

type DataSet struct {
	energy []float64
	sig []float64
	thermalSig float64
	epicadmiumSig []float64
}
var thermalEn float64 = 0.0253
var epicadmiumEn float64 = 0.4

func main() {
	handlingFile("inp.txt")
	// fmt.Println(data)
	url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9022734&req=2566&PenSectID=13660487&json"
	bytesBody := request(&url)
	var results map[string]interface{}
	err := json.Unmarshal(bytesBody, &results) // parsing json
	fmt.Println(err)
	pts := results["datasets"].([]interface{})[0].(map[string]interface{})["pts"].([]interface{}) // accessing pts map as a slice
	fmt.Printf("%T\n", pts)
	ds := &DataSet{}
	fulfilingStruct(pts, ds)
	fmt.Println(ds.thermalSig)
}

func handlingFile(name string) {
	file, _ := os.Open(name)
	scan := bufio.NewScanner(file)
	var linesArray [][]float64
	var incr int = 0
	for scan.Scan(){
		linesArray[] := strings.Split(scan.Text(), "  ")
		fmt.Println(splitting)
		incr++
	}
}

func fulfilingStruct(intr []interface{}, DS *DataSet) {
	// intrPrefix := intr.(map[string]interface{})
	for _, v := range intr {
		// fmt.Println(v.(map[string]interface{})["E"])
		DS.energy = append(DS.energy, v.(map[string]interface{})["E"].(float64))
		DS.sig = append(DS.sig, v.(map[string]interface{})["Sig"].(float64))
		if v.(map[string]interface{})["E"].(float64) == thermalEn {
			DS.thermalSig = v.(map[string]interface{})["Sig"].(float64)
		} else if v.(map[string]interface{})["E"].(float64) > epicadmiumEn {
			DS.epicadmiumSig = append(DS.epicadmiumSig, v.(map[string]interface{})["E"].(float64))
		}
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