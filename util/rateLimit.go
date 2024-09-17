package util

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Global cache map for recent users, used to track the amount of requests
// they have made in an alotted time slot (timeSlotSeconds).
var IpRegistry = NewThreadSafeMap[string, int]()

const (
	// The maximum request for a user per timeSlotSeconds
	maximumRequests = 20

	// The time slot after which request counters will be zeroed for users
	timeSlotSeconds = 60

	// The time slot after which we clear the gathered keys. This is done so
	// we don't take up a lot of memory eventually, since this map is going to grow
	// continuosly, but some users may not log in for a long time.
	cacheClearMinutes = 1440
)

// This function will act as a reset switch, checking every timeSlotSeconds
// and resetting the counters for each user. When cacheClearMinutes passes
// this function will clear all gathered keys in the map.
func RateLimit() {
	resetTicker := time.NewTicker(timeSlotSeconds * time.Second)
	clearTicker := time.NewTicker((cacheClearMinutes * 60) * time.Second)

	for {
		select {
		case t := <-resetTicker.C:
			log.Println("Resetting all user request counters at", t)
			IpRegistry.SetAll(0)
		case t := <-clearTicker.C:
			log.Println("Clearing all gathered user remote addresses at", t)
			IpRegistry.RemoveAllElements()
		}
	}
}

func CheckForRateLimit(request *http.Request) error {
	remoteAddr := request.RemoteAddr
	counter, exists := IpRegistry.Get(remoteAddr)

	if !exists {
		IpRegistry.Set(remoteAddr, 1)
		return nil
	}

	if counter < maximumRequests {
		IpRegistry.Set(remoteAddr, counter+1)
		return nil
	}

	return fmt.Errorf("Reached maximum allowed requests. Try again in %d seconds", timeSlotSeconds)
}
