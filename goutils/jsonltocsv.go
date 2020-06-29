package goutils

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

func JsonToCsv(inFile string, fileBytes *[]byte) {
	f, err := os.Create(inFile)
	if err != nil {
		log.Println(fmt.Sprintf("status=out_file_create_fail file=%s", inFile))
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	header := "username,password\n"
	fmt.Fprint(w, header)
	scanner := bufio.NewScanner(bytes.NewReader(*fileBytes))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		textTrim := strings.Trim(scanner.Text(), "[\"\"]")
		textSplit := strings.Split(textTrim, "\",\"")
		if len(textSplit) == 2 {
			credString := textSplit[0] + "," + textSplit[1] + "\n"
			fmt.Fprint(w, credString)
		}
	}
	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}
}