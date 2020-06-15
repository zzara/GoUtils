package main

import (
  "flag"
  "goutils"
)

var flagFlattenJson = flag.String("flattenjson", "false", "Flatten JSON to an un-nested JSON")
var flagInput = flag.String("i", "false", "Input file")
var flagOutput = flag.String("o", "false", "Outputs the modified file")

func main() {
  if *flattenJson != "false" {
    goutils.FlattenJson(*flagInput, *flagOutput)
  }
}
