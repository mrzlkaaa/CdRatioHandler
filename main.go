package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"bufio"
	"strings"
	"strconv"
	// "log"
)

type DataSet struct {
	energy []float64
	sig []float64
	thermalSig float64
	epicadmiumSig []float64
	resonseI []float64
	sumI float64
}
var thermalEn float64 = 0.0253
var epicadmiumEn float64 = 0.4

func main() {
	array := handlingFile("inp.txt")
	fmt.Println(array)
	url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9022734&req=2566&PenSectID=13660487&json"
	bytesBody := request(&url)
	var results map[string]interface{}
	err := json.Unmarshal(bytesBody, &results) // parsing json
	fmt.Println(err)
	pts := results["datasets"].([]interface{})[0].(map[string]interface{})["pts"].([]interface{}) // accessing pts map as a slice
	fmt.Printf("%T\n", pts)
	ds := &DataSet{}
	fulfilingStruct(pts, ds)
}

func convertStrings(array []string) []float64{
	floated := []float64{}
	for _, v := range array {
		floatv, _ := strconv.ParseFloat(v, 64)
		floated = append(floated, floatv)
	}
	return floated
}

func handlingFile(name string) [][]float64{
	file, _ := os.Open(name)
	scan := bufio.NewScanner(file)
	var linesArray [][]float64
	for scan.Scan(){
		if datArray := strings.Split(scan.Text(), "   "); len(datArray) > 1 {
			floatArray := convertStrings(datArray)
			// fmt.Println(floatArray)
			linesArray = append(linesArray, floatArray)
		}
	}
	return linesArray
}

func fulfilingStruct(intr []interface{}, DS *DataSet) {
	for _, v := range intr {
		DS.energy = append(DS.energy, v.(map[string]interface{})["E"].(float64))
		DS.sig = append(DS.sig, v.(map[string]interface{})["Sig"].(float64))
		if v.(map[string]interface{})["E"].(float64) == thermalEn {
			DS.thermalSig = v.(map[string]interface{})["Sig"].(float64)
		} else if v.(map[string]interface{})["E"].(float64) > epicadmiumEn {
			DS.epicadmiumSig = append(DS.epicadmiumSig, v.(map[string]interface{})["E"].(float64))
		}
	}
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println("Finished looping in sigms")
		}
	}()
	for i:=0; i <= len(DS.sig); i++ {
		if DS.energy[i] > 0.5 {
			DS.resonseI = append(DS.resonseI, DS.sig[i]*(DS.energy[i+1]-DS.energy[i])/average(DS.energy[i+1],DS.energy[i]))
			DS.sumI += DS.resonseI[i]
			// fmt.Println(sumI)
		} else {
			DS.resonseI = append(DS.resonseI, 0)
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

func average(lower, upper float64) float64 {
	avr := (lower+upper)/2
	return avr
}