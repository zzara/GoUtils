package main

import (
	"flag"
)

var FlagFlattenJson = flag.String("flattenjson", "false", "Flatten JSON to an un-nested JSON")
var FlagTimeCount = flag.String("timecount", "false", "Count from one time to another using a back time and increment")
var FlagTimeCountBack = flag.String("back", "false", "Timecount, back time")
var FlagTimeCountIncr = flag.String("increment", "false", "Timecount, increment")
var FlagInput = flag.String("i", "false", "Input file")
var FlagOutput = flag.String("o", "false", "Outputs the modified file")
var FlagUrl = flag.String("url", "false", "Outputs the modified file")