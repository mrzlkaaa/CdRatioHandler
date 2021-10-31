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
	energy []float64
	sig []float64
	thermalSig float64
	epicadmiumEn []float64
	epicadmiumSig []float64
	resonseSig []float64
	sumI float64
}
var thermalEn float64 = 0.0253
var epicadmiumEn float64 = 0.5 // rename it

func main() {
	start := time.Now()
	array := handlingFile("inp.txt")
	fmt.Println(array)
	url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9017305&req=11650&PenSectID=13655316&json"
	bytesBody := request(&url)
	var results map[string]interface{}
	err := json.Unmarshal(bytesBody, &results) // parsing json
	fmt.Println(err)
	pts := results["datasets"].([]interface{})[0].(map[string]interface{})["pts"].([]interface{}) // accessing pts map as a slice
	fmt.Printf("%T\n", pts)
	ds := &DataSet{}
	fulfilingStruct(pts, ds)
	cdRatio := cdComputation(array, *ds)
	fmt.Println(ds.sumI, ds.thermalSig, cdRatio)
	fmt.Println(time.Since(start))

}

func cdComputation(inpArray [][]float64, DS DataSet) float64 {
	multiSum := 0.0
	thermals := inpArray[0][2]*DS.thermalSig
	fmt.Println(thermals)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println("Finished looping in array in sigms")
			ratio := thermals/multiSum+1
			fmt.Println(ratio)
			// return ratio
		}
	}()
	for _, v1 := range inpArray {
		// fmt.Println(v1)
		lowerL, upperL, flux := v1[0], v1[1], v1[2]
		fmt.Println(lowerL, upperL)
		for j := 0; j <= len(DS.epicadmiumEn)-1; j++ {
			if lowerL < DS.epicadmiumEn[j] && DS.epicadmiumEn[j] < upperL{
				// fmt.Println(multiSum)
				multiSum += flux*DS.epicadmiumSig[j]*(DS.epicadmiumEn[j+1]-DS.epicadmiumEn[j])/average(DS.epicadmiumEn[j+1],DS.epicadmiumEn[j])
			} else {
				continue
			}
		}
	}
	ratio := thermals/multiSum-1
	return ratio
}

func convertStrings(array []string) []float64{
	floated := []float64{}
	// floatv := 0.0
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

func handlingFile(name string) [][]float64{
	file, _ := os.Open(name)
	scan := bufio.NewScanner(file)
	var linesArray [][]float64
	for scan.Scan(){
		if datArray := strings.Split(scan.Text(), "\t"); len(datArray) > 1 {
			floatArray := convertStrings(datArray)
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
			DS.epicadmiumSig = append(DS.epicadmiumSig, v.(map[string]interface{})["Sig"].(float64))
			DS.epicadmiumEn = append(DS.epicadmiumEn, v.(map[string]interface{})["E"].(float64))
		}
	}
	defer func() {
		if err := recover(); err != nil {
			// fmt.Println(err)
			fmt.Println("Finished looping in sigms")
			DS.resonseSig = append(DS.resonseSig, 0)
			fmt.Println(len(DS.epicadmiumEn), len(DS.resonseSig))
		}
	}()
	for i:=0; i <= len(DS.epicadmiumEn); i++ {
		DS.resonseSig = append(DS.resonseSig, DS.epicadmiumSig[i]*(DS.epicadmiumEn[i+1]-DS.epicadmiumEn[i])/average(DS.epicadmiumEn[i+1],DS.epicadmiumEn[i]))
		// resonseI := 
		DS.sumI += DS.resonseSig[i]
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