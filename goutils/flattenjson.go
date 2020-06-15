package goutils

// Takes a properly formatted JSON file and flattens it to a JSON line file

import (
	"os"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"time"
	"strings"
	"strconv"
)

var OUTDIR string
var JSONFILE string

func FlattenJson(flagInput string, flagOutput string) {
	JSONFILE = flagInput
	OUTDIR = flagOutput
	if len(JSONFILE) <= 0 {
		panic(fmt.Sprint("execution requires an input file: none found"))
	}
	outFile := prepareFile()
	defer outFile.Close()
	unmarshalledJson := jsonLoader(JSONFILE)
	for _, jsonObject := range unmarshalledJson {
		mappedJson, _ := jsonObject.(map[string]interface{})
		flattenedJson := flattener(mappedJson, "")
		writeJsonToFile(outFile, flattenedJson)
	}
	return
}

// Add a value to a key
func addValue(data interface{}, key string, value interface{}) {
    v, _ := data.(map[string]interface{})
    v[key] = value
    data = v
}

// Main JSON flattening recursion function
func flattener(unmarshalledJson map[string]interface{}, name string) map[string]interface{} {
	flattenedJson := make(map[string]interface{})
    for key, value := range unmarshalledJson {
		var key_name string
		if len(name) > 0 {
			key_name = fmt.Sprintf("%s_%s", name, key)
		} else {
			key_name = fmt.Sprintf("%s", key)
		}
        switch valueSwitch := value.(type) {
        case map[string]interface{}:
            for key_sub, value_sub := range flattener(valueSwitch, key_name) {
                addValue(flattenedJson, key_sub, value_sub)
            }
        case string:
			addValue(flattenedJson, key_name, fmt.Sprintf(strings.Replace(strings.Replace(valueSwitch, "\n", " ", -1), "\r", " ", -1)))
		case float64:
			addValue(flattenedJson, key_name, fmt.Sprintf(strconv.FormatFloat(valueSwitch, 'f', -1, 64)))
        case bool:
			addValue(flattenedJson, key_name, fmt.Sprintf(strconv.FormatBool(valueSwitch)))
        case nil:
			addValue(flattenedJson, key_name, fmt.Sprintf("nil"))
        }
	}
	return flattenedJson
}

// Loads the JSON file
func jsonLoader(filePath string) []interface{} {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("message=failed_to_open_query_file error=", err)
	} else {
		fmt.Println("status=successfully_opened_query_json_file message=", filePath)
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)
	var result []interface{}
	json.Unmarshal(byteValue, &result)
	return result
}

// Writes a JSON line to a file
func writeJsonToFile(outFile *os.File, flattenedJson map[string]interface{}) {
	marshalledJson, err := json.Marshal(flattenedJson)
	if err != nil {
		fmt.Println("status=failed_body_parse error=", err)
	}
	outFile.Write(marshalledJson)
	outFile.WriteString("\n")
}

// Creates output directory and new output file
func prepareFile() *os.File {
	if _, err := os.Stat(OUTDIR); os.IsNotExist(err) {
		err = os.MkdirAll(OUTDIR, 0755)
		if err != nil {
				panic(err)
		}
	}
	fileName := fmt.Sprintf("FJ%d.txt", time.Now().UnixNano())
	fileLocation := fmt.Sprintf("%s/%s", OUTDIR, fileName)
	outFile, err := os.Create(fileLocation)
    if err != nil {
        fmt.Println(err)
	}
	return outFile
}
