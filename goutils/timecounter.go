package goutils

import (
	"time"
	"log"
	"strconv"
)

func TimeCount(flagBackTimeStr string, offsetTimeStr string) {
	backTimeInt, _ := strconv.Atoi(flagBackTimeStr)
	offsetTimeInt, _ := strconv.Atoi(offsetTimeStr)
	fromTime, toTime, nextOffset := generateSearchTime(backTimeInt, offsetTimeInt)
	for fromTime.Before(toTime) {
		log.Printf("From Time: %s, Next Offset Time: %s, End Time: %s", fromTime, nextOffset, toTime)
		nextFromTime := incrementTime(fromTime, offsetTimeInt)
		nextOffset = incrementTime(nextOffset, offsetTimeInt)
		fromTime = nextFromTime
	}	
}

// Create the timestamp query parameters
// BACKTIME specifies the temporal assignment
func generateSearchTime(backTimeInt int, offsetTime int) (time.Time, time.Time, time.Time) {
	backTime := time.Duration(-1) * time.Duration(backTimeInt) * time.Hour
	toTime := time.Now().UTC()
	fromTime := toTime.Add(backTime)
	offset := incrementTime(fromTime, offsetTime)
	return fromTime, toTime, offset
}

// Increment time offset by +10 minutes
func incrementTime(timeString time.Time, increment int) time.Time {
	timeStamp := timeString.Add(time.Duration(increment) * time.Minute)
	return timeStamp
}