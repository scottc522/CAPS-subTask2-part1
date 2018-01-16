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
	"time"
)

//Structure containing an x and y cordinate
type dartCoords struct {
	x, y float64
}

//Real value of pi
const Pi = 3.14159

//number of dartboards(threads)
const numDartBoards = 4

//Number of times the routine is carried out
const dartsToThrow = 105555555

//Stores the number of hits on the dart board
var totalCount int

//Stores the number of itterations carried
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
	rand.Seed(999) // Seed the random number generator (constant seed makes a repeatable program)

	// Open a file to write results to.
	PiResults, fError := os.OpenFile("PiResults.txt", os.O_APPEND|os.O_CREATE, 0666) // create file if not existing, and append to it. 666 = permission bits rw-/rw-/rw-
	checkFileOK(fError)
	defer PiResults.Close() // defer means do this last, just before surrounding block terminates (ie 'main')

	// Set up required channels
	workchan := make([]chan int, numDartBoards) // Create arrays of integer channels. As many channels as there are dartboards
	hitResult := make([]chan int, numDartBoards)

	for i := range workchan { // Now set them up.
		workchan[i] = make(chan int)
		hitResult[i] = make(chan int)
	}

	//******** BEGIN TIMING HERE ***************************************************
	startTimer := time.Now()
	//work is the amount of work that needs to be done by each dart board.
	var work = dartsToThrow / (dartsToThrow / 1000)
	// Dart board processors begin in parallel
	for i := 0; i < numDartBoards; i++ {

		go DartBoard(i, workchan[i], hitResult[i])
	}

	// farmer/dart thrower process /////////////////////////////////////////

	for dartsThrown = 0; dartsThrown < dartsToThrow; {
	myloop:
		for i := range workchan { //checks each dart board of a result
			select {
			// DART BOARD i - use a non-blocking check by using default: case
			case count := <-hitResult[i]: // get 0 or 1 from dart board

				workchan[i] <- work // Send more work to the dartboard that has finished.
				//IF darts thrown is equal to itsel and a set of work and is greater than or equal than the
				//darts that needed to be thrown then break the loop
				if dartsThrown = dartsThrown + work; dartsThrown >= dartsToThrow {
					break myloop // and break out if exceeded

				}
				//at the count returned from the dartboard to the total count
				totalCount += count

			}
		}
	}

	//******** END TIMING HERE ***************************************************
	//total time taken to get result
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
	//Create new random object that is exclusive to this instance of the function
	rn := rand.New(rand.NewSource(int64(id)))

	hitOrMiss <- 0 // Trigger the sending of work by sending 0.
	var hitcount int
	for { // do forever
		hitcount = 0      //initialise hit counter to 0
		count := <-myWork //recieves the amount of work needed
		for i := 0; i < count; i++ {

			//create new dart and set its co-ordinates
			var dart dartCoords
			dart.x = rn.Float64()
			dart.y = rn.Float64()

			//record whether hit or miss
			hit := 1 - int(dart.x*dart.x+dart.y*dart.y)
			//add it to the running total
			hitcount = hitcount + hit
		}

		//send the hit count back to the main function via channel
		hitOrMiss <- hitcount

	}
}
