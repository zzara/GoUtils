package goutils

// Package for extracting text and urls from files
// File Types: text, text/html, rtf
//			   pdf, doc/docx, xls/xlsx, ppt/pptx
//			   gzip, gzip/bz2, tar, zip
// Limited support for all other file types (file strings only)
// V1.0

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
    "net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/h2non/filetype"
	"github.com/unidoc/unipdf/contentstream"
	"github.com/unidoc/unipdf/core"
	"github.com/unidoc/unipdf/model"
)

// Struct of a file and its properties
type File struct {
	fileBytes		[]byte
	fileName		string
	fileType		string
	fileExtension	string
	fileStrings		[]string
	fileParseMethod	func(file *File) ([]string, error)
}

// Map of file type strings handled by IX to parser functions
var FileTypes = map[string]func(file *File) ([]string, error) {
	"application/msword": extractStrings,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": extractZip,
	"application/vnd.ms-excel": extractZip,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": extractZip,
	"application/vnd.ms-powerpoint": extractZip,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": extractZip,
	"application/pdf": extractPdf,
	"text/html": extractStrings,
	"application/gzip": extractGzip,
	"application/x-bzip2": extractGzip,
	"application/zip": extractZip,
}

// Parse a file, return the struct of the parsed file
func LoadFile(fileBytes []byte, fileName string) *File {
	// Detect file type
	fileType, fileExtenstion, fileParseMethod, _ := detectFileInfo(fileBytes, fileName)
	return &File{
		fileBytes:		fileBytes,
		fileName:		fileName,
		fileType:		fileType,
		fileExtension:	fileExtenstion,
		fileParseMethod:fileParseMethod,
	}
}

// Detect file information
func detectFileInfo(fileBytes []byte, fileName string) (string, string, func(file *File) ([]string, error), error) {
	// Use filetype package to parse, with http package as a fallback
	kind, _ := filetype.Match(fileBytes)
	var fileExtenstion string
	var fileType string
	if kind == filetype.Unknown {
		fileType = http.DetectContentType(fileBytes)
		fileExtenstion = "unknown"
	} else {
		fileType = string(kind.MIME.Value)
		fileExtenstion = string(kind.Extension)
	}
	// Determine the parsing method to use based on file type
	var fileParseMethod func(file *File) ([]string, error)
	for fileTypeName, function := range FileTypes {
		if strings.HasPrefix(fileType, fileTypeName) {
			fileParseMethod = function
			break
		}
	}
	// Assign default string parser for undetected file types
	if fileParseMethod == nil {
		fileParseMethod = extractStrings
	}
	return fileType, fileExtenstion, fileParseMethod, nil
}

// Return file name string
func (file *File) GetFileName() string {
	return file.fileName
}

// Return file type string
func (file *File) GetFileType() string {
	return file.fileType
}

// Return file extension string
func (file *File) GetFileExtension() string {
	return file.fileExtension
}

// Return file bytes
func (file *File) GetFileBytes() []byte {
	return file.fileBytes
}

// Return file strings slice
func (file *File) GetFileStrings() []string {
	var stringSlice []string
	for _, str := range file.fileStrings {
		stringSlice = append(stringSlice, fmt.Sprintf("%+q", str))
	}
	return stringSlice
}

// Return file string
func (file *File) GetFileString() string {
	var singleString string
	for _, str := range file.fileStrings {
		singleString = singleString + str
	}
	return fmt.Sprintf("%+q", singleString)
}

// Parse files using the detected file parse method from FileTypes struct
func (file *File) Parse() ([]string, error) {
	var err error
	file.fileStrings, err = file.fileParseMethod(file)
	if err != nil {
		return nil, err
	}
	return file.fileStrings, nil
}

// Extract strings from microsoft documents and zip archives
func extractZip(file *File) ([]string, error) {
    zipReader, err := zip.NewReader(bytes.NewReader(file.fileBytes), int64(len(file.fileBytes)))
    if err != nil {
        return nil, err
	}
	var unzippedFileStrings []string
    // Read all the files from archive
    for _, zipFile := range zipReader.File {
		unzippedFile, err := zipFile.Open()
		if err != nil {
			continue
		}
		defer unzippedFile.Close()
		uzFileBytes, err := ioutil.ReadAll(unzippedFile)
		if err != nil {
			return nil, err
		}
		extractStrings, err := extractStrings(&File{fileBytes: uzFileBytes})
		if err != nil {
			return nil, err
		}
		for _, ts := range extractStrings {
			unzippedFileStrings = append(unzippedFileStrings, ts) 
		}
	}
	return unzippedFileStrings, nil
}

// Decompresses gzip archives
func extractGzip(file *File) ([]string, error) {
	var unzippedFileStrings []string
	bytesReader := bytes.NewReader(file.fileBytes)
	var tarReader *tar.Reader
	if file.fileExtension == "bz2" {
		bz2Reader := bzip2.NewReader(bytesReader)
		tarReader = tar.NewReader(bz2Reader)
	} else if file.fileExtension == "gz" {
		gzReader, err := gzip.NewReader(bytesReader)
		if err != nil {
			return nil, err
		}
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(bytesReader)
	}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			bytesBuffer := new(bytes.Buffer)
			bytesBuffer.ReadFrom(tarReader)
			bufferBytes := bytesBuffer.Bytes()
			tarStrings, err := extractStrings(&File{fileBytes: bufferBytes})
			if err != nil {
				return nil, err
			}
			for _, ts := range tarStrings {
				unzippedFileStrings = append(unzippedFileStrings, ts) 
			}
		default:
			continue
		}
	}
	return unzippedFileStrings, nil
}

// Extract PDF page text and resources
// 1. Parse content streams for visible strings
// 2. Parse object streams for resource dependencies, get strings
// 3. Parse raw file for strings
func extractPdf(file *File) ([]string, error) {
	var outputStrings []string
	// Content stream visible text parse
	outVisibleText, err := extractPdfVisibleText(file.fileBytes)
	if err != nil {
		log.Printf("MODULE=extractPdfVisibleText ERROR=%s", err)
	} else {
		for _, text := range outVisibleText {
			outputStrings = append(outputStrings, text)
		}
	}
	// Object stream resource dependency parse
	outStreamText, err := extractPdfObjectStreams(file.fileBytes)
	if err != nil {
		log.Printf("MODULE=extractPdfObjectStreams ERROR=%s", err)
	} else {
		for _, stream := range outStreamText {
			outputStrings = append(outputStrings, stream)
		}
	}
	// Extract strings from whole raw file
	rawStrings, err := extractStrings(file)
	if err != nil {
		log.Printf("MODULE=extractStrings ERROR=%s", err)
	} else {
		for _, str := range rawStrings {
			outputStrings = append(outputStrings, str)
		}
	}
	return outputStrings, nil
}

// Extract strings from raw bytes
func extractStrings(file *File) ([]string, error) {
	byteString := string(file.fileBytes)
	var cleanStrings []string
	if !utf8.ValidString(byteString) {
		v := make([]rune, 0, len(byteString))
		for i, r := range byteString {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(byteString[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		byteString = string(v)
	}
	if len(byteString) > 0 {
		cleanStrings = append(cleanStrings, byteString)
	}
	return cleanStrings, nil
}

// Extract visible text from content streams on PDF pages
func extractPdfVisibleText(file []byte) ([]string, error) {
	var outText []string
	bytesReader := bytes.NewReader(file)
	pdfReader, err := model.NewPdfReader(bytesReader)
	if err != nil {
		return nil, err
	}
	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return nil, err
	}
	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		if err != nil {
			return nil, err
		}
	}
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return nil, err
	}
	// Visible PDF to text extract, loop
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("MODULE=extractPdfVisibleText OPERATION=pdfReader.GetPage ERROR=%s", err)
			continue
		}
		contentStreams, err := page.GetAllContentStreams()
		if err != nil {
			log.Printf("MODULE=extractPdfVisibleText OPERATION=page.GetAllContentStreams ERROR=%s", err)
			continue
		} 
		cstreamParser := contentstream.NewContentStreamParser(contentStreams)
		operations, err := cstreamParser.Parse()
		if err != nil {
			log.Printf("MODULE=extractPdfVisibleText OPERATION=cstreamParser.Parse ERROR=%s", err)
			continue
			}
		var outString string
		var pdFont *model.PdfFont
		for _, op := range *operations {
			// Search font type object, set it to process Tj
			if op.Operand == "Tf" {
				switch op.Params[0].(type) {
				case *core.PdfObjectName:
					name := op.Params[0].(*core.PdfObjectName)
					fontObj, found := page.Resources.GetFontByName(*name)
					if found {
						pdFont, _ = model.NewPdfFontFromPdfObject(fontObj)
					}
				}
			} else if op.Operand == "Tj" {
				// Get text, process encodings for actual text
				charcodesBytes, ok := core.GetStringBytes(op.Params[0])
				if ok {
					if pdFont != nil {
						charcodes := pdFont.BytesToCharcodes(charcodesBytes)
						runes, _, _ := pdFont.CharcodesToUnicodeWithStats(charcodes)
						for _, r := range runes {
							if r == '\x00' {
								continue
							}
							outString += string(r)
						}
					}
				}
			}
		}
		outText = append(outText, outString)
	}
	return outText, nil
}

// Parse all PDF object streams, flatten and resolve values, return strings
func extractPdfObjectStreams(file []byte) ([]string, error) {
	var outText []string
	readSeeker := bytes.NewReader(file)
	pdfParser, err := core.NewParser(readSeeker)
	if err != nil {
		log.Printf("MODULE=extractPdfObjectStreams OPERATION=core.NewParser ERROR=%s", err)
		} else {
			// Parsed Object Streams
			for i := 1; i < 1000; i++ {
				numStream, err := pdfParser.LookupByNumber(i)
				numStreamString := fmt.Sprintf("%v", numStream)
				if err == nil {
					if numStreamString == "null" || i > 1000 {
						break
						} else {
							flatRef := core.FlattenObject(numStream)
							outText = append(outText, fmt.Sprintf("%v", flatRef))
						}
				} else {
					break
				}
			}
	}
	return outText, nil
}

// Extract and format urls, keeping only unique urls
func (file *File) UrlExtract() []string {
	var urls []string
	urlRe := regexp.MustCompile(`(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`)
	for _, str := range file.fileStrings {
		urlMatch := urlRe.FindAllString(str, -1)
		if urlMatch != nil {
			for _, url := range urlMatch {
				urls = append(urls, fmt.Sprintf("%s", url))
			}
		}
	}  
	urls = uniqueUrls(urls)
	return urls
}

// Take a string slice and filter out duplicate values
func uniqueUrls(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}