package main

import (
  "flag"
)

var flattenJson = flag.Bool("flattenjson", "false", "Flatten JSON to an un-nested JSON")
var output = flag.Bool("out", "false", "Outputs the modified file")

func main() {
  if flattenJson != "false" {
    flattenJson()
  }
}
