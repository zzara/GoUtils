package goutils

// Bz2 compressed results parser

import (
	"os"
	"fmt"
	"compress/bzip2"
	"bytes"
	"log"
	"archive/tar"
	"io/ioutil"
	"io"
)

var BZOUTDIR string
var BZFILE string

func Bz2Parser(flagInput string, flagOutput string) []string {
	var fileStrings []string
	BZFILE = flagInput
	BZOUTDIR = flagOutput
	f, err := os.Open(flagInput)
	if err != nil {
		log.Println("message=failed_to_open_query_file error=", err)
	} else {
		log.Println("status=successfully_opened_query_json_file message=", flagInput)
	}
	defer f.Close()
	bz2Bytes, _ := ioutil.ReadAll(f)
	// Create bytes reader
	bytesReader := bytes.NewReader(bz2Bytes)
	// Create a bzip2.reader
	bz2f := bzip2.NewReader(bytesReader)
	// Create a tarfile reader
	tarReader := tar.NewReader(bz2f)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(fmt.Sprintf("%s", err))
		} else {
			//name := header.Name
			switch header.Typeflag {
			case tar.TypeDir:
				continue
			case tar.TypeReg:
				buf := new(bytes.Buffer)
				buf.ReadFrom(tarReader)
				newString := buf.String()
				fileStrings = append(fileStrings, newString)
			default:
				log.Println(fmt.Sprintf("function=Bz2Parser process=parse_tarfile message=could_not_parse_bz2.tar_file type=%b", header.Typeflag))
			}
		}
	}
	return fileStrings
}