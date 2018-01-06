// Pi on the Dart Board - Go Version with concurrency
// This version sends individual x,y, pairs to the workers = communication heavy
// A.ORAM Nov 2017
// Version 3. Same as V2TF but uses range and break to avoid too many darts, but back to round robin querying
// of dart boards.
package main

// Imported packages

import (
	"fmt" // for console I/O
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"
)

type dartCoords struct {
	x, y float64
}

const Pi = 3.14159
const numDartBoards = 4
const dartsToThrow = 10555555555

var totalCount int
var dartsThrown int

var n int // used to dump byte counts from file IO

// File OK check
func checkFileOK(ecode error) { // type 'error' is known to Go
	if ecode != nil {
		panic(ecode) // crash out of the program!
	}
}

//////////////////////////////////////////////////////////////////////////////////
//  Main program, create required channels, then start goroutines in parallel.
//////////////////////////////////////////////////////////////////////////////////

func main() {
	rand.Seed(999)                                   // Seed the random number generator (constant seed makes a repeatable program)
	fmt.Println("Max cores=", runtime.GOMAXPROCS(0)) // Set the max number of cores whilst seeing what it currently is.

	// Open a file to write results to.
	PiResults, fError := os.OpenFile("PiResults.txt", os.O_APPEND|os.O_CREATE, 0666) // create file if not existing, and append to it. 666 = permission bits rw-/rw-/rw-
	checkFileOK(fError)
	defer PiResults.Close() // defer means do this last, just before surrounding block terminates (ie 'main')

	// Set up required channels
	workchan := make([]chan int, numDartBoards) // Create arrays of channels
	hitResult := make([]chan int, numDartBoards)

	for i := range workchan { // Now set them up.
		workchan[i] = make(chan int)
		hitResult[i] = make(chan int)
	}

	// Now start the dart board workers in parallel.
	fmt.Println("\nStart Dart Board processes...")

	//******** BEGIN TIMING HERE ***************************************************
	startTimer := time.Now()
	var work = dartsToThrow / (dartsToThrow / 1000)
	for i := 0; i < numDartBoards; i++ {

		go DartBoard(i, workchan[i], hitResult[i])
	}

	// farmer/dart thrower process /////////////////////////////////////////

	for dartsThrown = 0; dartsThrown < dartsToThrow; {
	myloop:
		for i := range workchan { // Round robin check each dart board for a result (polling?)
			select {
			// DART BOARD i - use a non-blocking check by using default: case
			case count := <-hitResult[i]: // get 0 or 1 from dart board
				//fmt.Println(dartsThrown)
				workchan[i] <- work // immediately throw another dart its way

				//dartsThrown++         // keep a tally of darts thrown
				// (THIS WILL NOT BE EXACTLY = dartsToThrow = error? May be OK, more darts the merrier?)
				if dartsThrown = dartsThrown + work; dartsThrown >= dartsToThrow { // SOLUTION: increment tally and check for limit
					break myloop // and break out if exceeded

				}

				totalCount += count
			default: // if dart board not ready check next...
			}
		}
	}

	//******** END TIMING HERE ***************************************************
	elapsedTime := time.Since(startTimer)

	fmt.Println("Total number of darts thrown =", dartsThrown)
	fmt.Println("Final hit count =", totalCount)

	// do PI calculations and display/save results...
	var PiApprox = (4.0 * float64(totalCount)) / float64(dartsThrown)
	var PiError = math.Abs((Pi - PiApprox) * (100.0 / Pi))
	fmt.Println("\n \n", dartsThrown, totalCount)
	fmt.Println("\n\nPi approx   =", PiApprox, " using", dartsThrown, "darts")
	fmt.Println("\nPi actually =", Pi, "  Error =", PiError, "%")

	log.Println("Elapsed time = ", elapsedTime) // Main difference 'log' gives you is the time and date it logs too... handy

	n, fError = PiResults.WriteString(fmt.Sprintf("\n\nPi approx   = %f using %d darts", PiApprox, dartsThrown)) // Sprintf formats output as string
	n, fError = PiResults.WriteString(fmt.Sprintf("\n\nPi actually = %f Error = %f%%", Pi, PiError))             // to keep WriteString happy.
	n, fError = PiResults.WriteString(fmt.Sprintf("\n\nElapsed time = %v", elapsedTime))
	n, fError = PiResults.WriteString("\n________________________________________________________________________")
	checkFileOK(fError)

} // end of main /////////////////////////////////////////////////////////////////

//----------------------------------------------------------------------------------
// Each dart board figures out if the x,y coord is a hit or not and returns 1 or 0
//----------------------------------------------------------------------------------

func DartBoard(id int, myWork <-chan int, hitOrMiss chan<- int) {
	rn := rand.New(rand.NewSource(int64(id)))

	//fmt.Println("Dart board", id, "ready!")
	hitOrMiss <- 0 // kick off the process by sending a zero
	hitcount := 0
	for { // do forever
		hitcount = 0
		count := <-myWork //recieves the amount of work needed
		for i := 0; i < count; i++ {

			//fmt.Println("Dartboard", id, "SetOfWork")
			var dart dartCoords
			dart.x = rn.Float64()
			dart.y = rn.Float64()
			hit := 1 - int(dart.x*dart.x+dart.y*dart.y)
			//fmt.Println("hit or miss", hit)
			hitcount = hitcount + hit
		}

		//time.Sleep(time.Duration(rand.Intn(1)) * time.Microsecond) // artificially slow down

		hitOrMiss <- hitcount

	}
}
