package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	// "log"
)

type DataSet struct {
	name          string
	inputData     [][]float64
	energy        []float64
	sig           []float64
	thermalSig    float64
	epicadmiumEn  []float64
	epicadmiumSig []float64
	resonseSig    []float64
	sumI          float64
}

type Writer interface {
	HandleInput(string)
	WriteDataSet([]interface{})
	WriteFileTXT(string)
	WriteExcel(string)
	ComputateCd() float64
}

var thermalEn float64 = 0.0253
var epicadmiumThreshold float64 = 0.5 // rename it

func main() {
	start := time.Now()
	var ws Writer = NewDSInstance()
	ws.HandleInput("inp.txt")
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9027530&req=5197&PenSectID=13665283&json" //* AM-241(N,f)
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9027571&req=5197&PenSectID=13665324&json" // *AM-241(N,G)AM-242
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9027723&req=5203&PenSectID=13665476&json" // *AM-242(N,f)
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9027758&req=5203&PenSectID=13665511&json" // *AM-242(N,G)AM-243
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9028228&req=5207&PenSectID=13665981&json" //* CM-242(N,f)
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9028237&req=5207&PenSectID=13665990&json" //* CM-242(N,G)CM-243
	// url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9026180&req=5254&PenSectID=13663933&json" //* Np-238(N,G)Np-239
	url := "https://www-nds.iaea.org/exfor/servlet/E4sGetTabSect?SectID=9026222&req=5254&PenSectID=13663975&json" //* Np-238(N,f)
	bytesBody := request(&url)
	var results map[string]interface{}
	err := json.Unmarshal(bytesBody, &results) // parsing json
	fmt.Println(err)

	pts := results["datasets"].([]interface{})[0].(map[string]interface{})["pts"].([]interface{}) // accessing pts map as a slice
	fmt.Printf("%T\n", pts)
	ws.WriteDataSet(pts)
	ws.WriteExcel(results["datasets"].([]interface{})[0].(map[string]interface{})["REACTION"].(string))

	fmt.Println(ws.ComputateCd())
	fmt.Println(time.Since(start))
}

func (wr *DataSet) HandleInput(name string) {
	file, _ := os.Open(name)
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		if datArray := strings.Split(scan.Text(), "\t"); len(datArray) > 1 {
			floatArray := convertStrings(datArray)
			wr.inputData = append(wr.inputData, floatArray)
		}
	}
}

func (DS *DataSet) WriteFileTXT(name string) {
	f, err := os.Create(name + ".txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for i := 0; i < len(DS.energy); i++ {
		en := fmt.Sprintf("%f", DS.energy[i])
		sig := fmt.Sprintf("%f", DS.sig[i])
		writeFormat := fmt.Sprintf("%v    %v\n", en, sig)
		f.WriteString(writeFormat)
	}

}

func (DS *DataSet) WriteExcel(name string) {
	xlsx := excelize.NewFile()

	for i := 0; i < len(DS.energy); i++ {
		// en := fmt.Sprintf("%f", DS.energy[i])
		// sig := fmt.Sprintf("%f", DS.sig[i])
		// writeFormat := fmt.Sprintf("%v    %v\n", en, sig)
		col1 := fmt.Sprintf("A%v", i+1)
		col2 := fmt.Sprintf("B%v", i+1)
		xlsx.SetCellValue("Sheet1", col1, DS.energy[i])
		xlsx.SetCellValue("Sheet1", col2, DS.sig[i])
	}

	err := xlsx.SaveAs(name + ".xlsx")
	if err != nil {
		fmt.Println(err)
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
	for i := 0; i <= len(DS.epicadmiumEn); i++ {
		DS.resonseSig = append(DS.resonseSig, DS.epicadmiumSig[i]*(DS.epicadmiumEn[i+1]-DS.epicadmiumEn[i])/average(DS.epicadmiumEn[i+1], DS.epicadmiumEn[i]))
		DS.sumI += DS.resonseSig[i]
	}
}

func (DS *DataSet) ComputateCd() float64 {
	multiSum := 0.0
	thermals := DS.inputData[0][2] * DS.thermalSig
	for _, v1 := range DS.inputData {
		Sum := 0.0
		lowerL, upperL, flux := v1[0], v1[1], v1[2]
		for j := 0; j <= len(DS.epicadmiumEn)-1; j++ {
			if lowerL <= DS.epicadmiumEn[j] && DS.epicadmiumEn[j] < +upperL {
				Sum += DS.resonseSig[j]
			}
		}
		multiSum += flux * Sum
	}
	fmt.Println(thermals/multiSum + 1)
	return thermals/multiSum + 1
}

func NewDSInstance() *DataSet {
	return &DataSet{}
}

func convertStrings(array []string) []float64 {
	var floated []float64
	for ind, v := range array {
		floatv, _ := strconv.ParseFloat(v, 64)
		if ind < 2 {
			floatv = floatv * 1e6
		} else {
			floatv = floatv * 1.12e+13
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
	return (lower + upper) / 2
}
