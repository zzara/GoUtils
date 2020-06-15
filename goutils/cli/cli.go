package main

import (
  "flag"
  "goutils"
  "log"
)

func main() {
  flag.Parse()
  if *FlagFlattenJson != "false" {
    log.Println("Starting FlattenJson")
    goutils.FlattenJson(*FlagInput, *FlagOutput)
  } else if *FlagTimeCount != "false" {
    log.Println("Starting TimeCount")
    if *FlagTimeCountBack != "false" && *FlagTimeCountIncr != "false" {
      goutils.TimeCount(*FlagTimeCountBack, *FlagTimeCountIncr)
    }
  } else {
    log.Println("No input received")
    log.Println(*FlagFlattenJson)
    log.Println(*FlagTimeCount)
  }
}
