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
	"time"
	// "log"
)

type DataSet struct {
	inputData [][]float64
	energy []float64
	sig []float64
	thermalSig float64
	epicadmiumEn []float64
	epicadmiumSig []float64
	resonseSig []float64
	sumI float64
}

type Writer interface {
	HandleInput(string)
	WriteDataSet([]interface{})
	ComputateCd() float64

}
var thermalEn float64 = 0.0253
var epicadmiumThreshold float64 = 0.5 // rename it

func main() {
	start := time.Now()
	var ws Writer = NewDSInstance()
	ws.HandleInput("inp.txt")
	url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=2292099&req=11678&PenSectID=7664758&json"
	bytesBody := request(&url)
	var results map[string]interface{}
	err := json.Unmarshal(bytesBody, &results) // parsing json
	fmt.Println(err)
	pts := results["datasets"].([]interface{})[0].(map[string]interface{})["pts"].([]interface{}) // accessing pts map as a slice
	fmt.Printf("%T\n", pts)
	ws.WriteDataSet(pts)
	cdRatio := ws.ComputateCd()
	fmt.Println(cdRatio)
	fmt.Println(time.Since(start))
}

func (wr *DataSet) HandleInput(name string){
	file, _ := os.Open(name)
	scan := bufio.NewScanner(file)
	for scan.Scan(){
		if datArray := strings.Split(scan.Text(), "\t"); len(datArray) > 1 {
			floatArray := convertStrings(datArray)
			wr.inputData = append(wr.inputData, floatArray)
		}
	}
}

func (DS *DataSet) WriteDataSet(intr []interface{}) {
	for _, v := range intr {
		DS.energy = append(DS.energy, v.(map[string]interface{})["E"].(float64))
		DS.sig = append(DS.sig, v.(map[string]interface{})["Sig"].(float64))
		if v.(map[string]interface{})["E"].(float64) == thermalEn {
			DS.thermalSig = v.(map[string]interface{})["Sig"].(float64)
		} else if v.(map[string]interface{})["E"].(float64) > epicadmiumThreshold {
			DS.epicadmiumSig = append(DS.epicadmiumSig, v.(map[string]interface{})["Sig"].(float64))
			DS.epicadmiumEn = append(DS.epicadmiumEn, v.(map[string]interface{})["E"].(float64))
		}
	}
	defer func() {
		if err := recover(); err != nil {
			// fmt.Println(err)
			fmt.Println("Finished looping in sigms")
			DS.resonseSig = append(DS.resonseSig, 0)
		}
	}()
	for i:=0; i <= len(DS.epicadmiumEn); i++ {
		DS.resonseSig = append(DS.resonseSig, DS.epicadmiumSig[i]*(DS.epicadmiumEn[i+1]-DS.epicadmiumEn[i])/average(DS.epicadmiumEn[i+1],DS.epicadmiumEn[i]))
		DS.sumI += DS.resonseSig[i]
	}
}

func (DS *DataSet) ComputateCd() float64 {
	multiSum := 0.0
	thermals := DS.inputData[0][2]*DS.thermalSig
	for _, v1 := range DS.inputData {
		Sum := 0.0
		lowerL, upperL, flux := v1[0], v1[1], v1[2]
		for j := 0; j <= len(DS.epicadmiumEn)-1; j++ {
			if lowerL <= DS.epicadmiumEn[j] && DS.epicadmiumEn[j] <+ upperL{
				Sum += DS.resonseSig[j]
			}
		}
		multiSum += flux*Sum
	}
	fmt.Println(thermals/multiSum+1)
	return thermals/multiSum+1
}

func NewDSInstance() *DataSet {
	return &DataSet{}
}

func convertStrings(array []string) []float64{
	var floated []float64
	for ind, v := range array {
		floatv, _ := strconv.ParseFloat(v, 64)
		if ind < 2 {
			floatv = floatv*1E6
		} else {
			floatv = floatv*1.12E+13
		}
		floated = append(floated, floatv)
	}
	return floated
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
	return (lower+upper)/2
}
