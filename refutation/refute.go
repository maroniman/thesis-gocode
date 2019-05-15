package main
// VERSION WITH OLD COINCIDENCE PROCESSING
import  (
	"fmt"
	"flag"
	"sort"
	"os"
	"time"
	"strconv"
)

type Coset struct {
	Id int
	Live bool
	Transitions []int
}

type ScanCommand struct {
	StartingCoset int
	Word []int
}

type Coincidence struct {
	Old int
	New int
}

type EqvClass struct {
	Live bool
	Parent int
	Children []int
}

type InitializationVector struct {
	N int
	M int
	Seed int
	ConservativeStrategy bool //DEPRECATED
	RunPlanted bool
	NumberOfRuns int
	StrictlyDifferentConstraints bool
}

var footballCosets [][]int = [][]int{[]int{2,3,4,4},
	[]int{ 5,   1,   6,   6 },
	[]int{   1,   7,   8,   8 },
	[]int{   9,  10,   1,   1 },
	[]int{   7,   2,  11,  11 },
	[]int{  12,  13,   2,   2 },
	[]int{   3,   5,  14,  14 },
	[]int{  15,  16,   3,   3 },
	[]int{  17,   4,  16,  16 },
	[]int{   4,  18,  12,  12 },
	[]int{  19,  20,   5,   5 },
	[]int{  21,   6,  10,  10 },
	[]int{   6,  22,  19,  19 },
	[]int{  23,  24,   7,   7 },
	[]int{  25,   8,  24,  24 },
	[]int{   8,  26,   9,   9 },
	[]int{  18,   9,  27,  27 },
	[]int{  10,  17,  28,  28 },
	[]int{  29,  11,  13,  13 },
	[]int{  11,  30,  23,  23 },
	[]int{  22,  12,  31,  31 },
	[]int{  13,  21,  32,  32 },
	[]int{  33,  14,  20,  20 },
	[]int{  14,  34,  15,  15 },
	[]int{  26,  15,  35,  35 },
	[]int{  16,  25,  36,  36 },
	[]int{  36,  37,  17,  17 },
	[]int{  38,  31,  18,  18 },
	[]int{  30,  19,  39,  39 },
	[]int{  20,  29,  40,  40 },
	[]int{  28,  41,  21,  21 },
	[]int{  42,  39,  22,  22 },
	[]int{  34,  23,  43,  43 },
	[]int{  24,  33,  44,  44 },
	[]int{  44,  45,  25,  25 },
	[]int{  46,  27,  26,  26 },
	[]int{  27,  47,  38,  38 },
	[]int{  48,  28,  37,  37 },
	[]int{  32,  49,  29,  29 },
	[]int{  50,  43,  30,  30 },
	[]int{  31,  48,  42,  42 },
	[]int{  51,  32,  41,  41 },
	[]int{  40,  52,  33,  33 },
	[]int{  53,  35,  34,  34 },
	[]int{  35,  54,  46,  46 },
	[]int{  47,  36,  45,  45 },
	[]int{  37,  46,  55,  55 },
	[]int{  41,  38,  56,  56 },
	[]int{  39,  51,  50,  50 },
	[]int{  57,  40,  49,  49 },
	[]int{  49,  42,  58,  58 },
	[]int{  43,  57,  53,  53 },
	[]int{  54,  44,  52,  52 },
	[]int{  45,  53,  59,  59 },
	[]int{  59,  56,  47,  47 },
	[]int{  55,  58,  48,  48 },
	[]int{  52,  50,  60,  60 },
	[]int{  56,  60,  51,  51 },
	[]int{  60,  55,  54,  54 },
	[]int{  58,  59,  57,  57 } }

var initVector InitializationVector = InitializationVector{-1,-1,-1, false, false, 1, false}

var cosetTable []*Coset
var scanQueue []ScanCommand
var coincidenceQueue []Coincidence
var eqvClasses []EqvClass
var deadCosets []int

var groupRelations [][]int
var constraintRelations [][]int
var reverseConstraintRelations [][]int
var numberOfGenerators int
var cosetTableWidth int

func main() {	
	initializeInitVector()
	
	approach7() //conservative scanning strategy, as in thesis
	
	//approach1() // initial idea with Chris, 6 phases
	//approach2() // recursively scan constraints infinitely
	//approach3() // scan constraints once
	//approach4() // scan group once, then constraints once and count group
	//approach5() // ~scan constraints once, then group once and count
	//approach6() // scan constraints recursively
}

func approach7() {
	fmt.Println("n=", initVector.N, "m=", initVector.M)
	unplantedFoundWithoutGuessing := 0
	unplantedFoundWithGuessing := 0
	for run := 0; run < initVector.NumberOfRuns; run++ {
		fmt.Println(" ")
		startTime := time.Now()
		oldLength := 1
		newLength := 1

		if(initVector.RunPlanted) {
			startTime = time.Now()		
			initialize(true, run) // planted instance
			//fmt.Println("started planted run at ", time.Now())

			// construct FB0
			for i := 1; i<63; i++ {	
				for _, relation := range groupRelations {
					pushScan(ScanCommand{i, relation})	
				}
			}
			scanAll()
			oldLength = 1
			newLength = len(cosetTable)
		
			for i := oldLength; i<newLength; i++ {
				for _, relation := range constraintRelations {
					pushScan(ScanCommand{i, relation})
				}
			}
			scanAll()

			oldLength = newLength
			newLength = len(cosetTable)
			//fmt.Println(oldLength, "---", newLength)

			for i := oldLength; i<newLength; i++ {
				scannedFB := false
				for j, relation := range constraintRelations {
					reverseRel := reverseConstraintRelations[j]
					// NEW Conservative scanning: Decide here what to scan
					conservativeCondition1 := cosetTable[i].Transitions[getColumnIndex(relation[0])] > 0 && cosetTable[i].Transitions[getColumnIndex(relation[0])] < 66
					conservativeCondition2 := cosetTable[i].Transitions[getColumnIndex(reverseRel[len(reverseRel)-1])+1] > 0 && cosetTable[i].Transitions[getColumnIndex(reverseRel[len(reverseRel)-1])+1] < 66				
					if conservativeCondition1 || conservativeCondition2 {
						if !scannedFB {
							scannedFB = true
							addFootball(i)
						}
						if  conservativeCondition1 {
							pushScan(ScanCommand{i, relation})
						} 
						if conservativeCondition2 {
							pushScan(ScanCommand{i, reverseRel})
						}
					}
				}
			}
			//fmt.Println(len(scanQueue),"to scan")
			scanAll()
		
			elapsed := time.Since(startTime)
			fmt.Println("PLANTED:   (s=",initVector.Seed+run,"), alive cosets:", numberOfLiveCosets(), ", used a total of", len(cosetTable), "cosets --- in", elapsed)
			
			fmt.Println("quit with 0. Continue by entering ID assignment \"guess\".")
			var ida int
			_, err := fmt.Scanf("%d", &ida)
			check(err)
			if ida > 0 {
				relationToScan := []int{ida}
				for i:=1;i<len(cosetTable);i++ {
					if cosetTable[i].Live {
						pushScan(ScanCommand{i, relationToScan})
					}
				}
				scanAll()
				alive := numberOfLiveCosets()
				fmt.Println("after guess alive cosets:", alive)
			}

			//printCosetTable(-1,-1)
			fmt.Println(" ")
		}

		// ------------------------------------------------------------------

		startTime = time.Now()
		
		initialize(false, run) // unplanted instance
		//fmt.Println("started run at ", startTime)
		// construct FB0
		for i := 1; i<63; i++ {	
			for _, relation := range groupRelations {
				pushScan(ScanCommand{i, relation})	
			}
		}
		scanAll()
		oldLength = 1
		newLength = len(cosetTable)
		
		for i := oldLength; i<newLength; i++ {
			for _, relation := range constraintRelations {
				pushScan(ScanCommand{i, relation})	
			}
		}
		scanAll()
		
		oldLength = newLength
		newLength = len(cosetTable)
		for i := oldLength; i<newLength; i++ {
			scannedFB := false
			for j, relation := range constraintRelations {
				reverseRel := reverseConstraintRelations[j]
				// Conservative scanning: Decide here what to scan
				conservativeCondition1 := cosetTable[i].Transitions[getColumnIndex(relation[0])] > 0 && cosetTable[i].Transitions[getColumnIndex(relation[0])] < 66
				conservativeCondition2 := cosetTable[i].Transitions[getColumnIndex(reverseRel[len(reverseRel)-1])+1] > 0 && cosetTable[i].Transitions[getColumnIndex(reverseRel[len(reverseRel)-1])+1] < 66				
				if conservativeCondition1 || conservativeCondition2 {
					if !scannedFB {
						scannedFB = true
						addFootball(i)
					}
					if  conservativeCondition1 {
						pushScan(ScanCommand{i, relation})
					} 
					if conservativeCondition2 {
						pushScan(ScanCommand{i, reverseRel})
					}
				}
			}
		}
		//fmt.Println(len(scanQueue),"to scan")
		if scanAll() {
			unplantedFoundWithoutGuessing++
		}
		
		elapsed := time.Since(startTime)
		fmt.Println("UNPLANTED: (s=",initVector.Seed+run,"), alive cosets:", numberOfLiveCosets(), ", used a total of", len(cosetTable), "cosets --- in", elapsed)
		//printCosetTable(-1,-1)
		ida := 6
		if ida > 0 {
			relationToScan := []int{ida}
			for i:=1;i<len(cosetTable);i++ {
				if cosetTable[i].Live {
					pushScan(ScanCommand{i, relationToScan})
				}
			}
			if scanAll() {
				unplantedFoundWithGuessing++
			}
			alive := numberOfLiveCosets()
			fmt.Println("after guess alive cosets:", alive, "(run",run,")")
		}

		//printCosetTable(-1,-1)
		fmt.Println(" ")
	}
	fmt.Println("unpl. found without guessing:", unplantedFoundWithoutGuessing, "/", initVector.NumberOfRuns)
	fmt.Println("unpl. found with guessing:   ", unplantedFoundWithGuessing, "/", initVector.NumberOfRuns)
	fmt.Println("--------------------------------------------")
}

func initialize(planted bool, run int) {
	filename := getFilename("../data/a5", initVector.N, initVector.M, initVector.Seed+run, planted, initVector.StrictlyDifferentConstraints)
	if _, err := os.Stat(filename); err == nil {
		// file exists
	} else {
		generateInstance(initVector.N,initVector.M,int64(initVector.Seed+run),initVector.StrictlyDifferentConstraints)
	}
	
	groupRelations, constraintRelations, reverseConstraintRelations, numberOfGenerators = readFile(filename)
	cosetTableWidth = numberOfGenerators*2

	cosetTable = make([]*Coset, 0, 10000)
	scanQueue = make([]ScanCommand, 0, 1000)
	coincidenceQueue = make([]Coincidence, 0, 1000)
	deadCosets = make([]int, 0, 100)
	
	defineCoset(-1, 0); // place dead coset at index 0
	cosetTable[0].Live = false
	
	defineCoset(-1, 0); // initial coset at index 1
}

func initializeInitVector() () {
	nPtr := flag.Int("n", -1, "number of variables")
	mPtr := flag.Int("m", -1, "number of constraints")
	seedPtr := flag.Int("seed", -1, "initial seed")
	consvPtr := flag.Bool("consv", false, "use conservative strategy while scanning phase 3")
	plantedPtr := flag.Bool("planted", false, "should planted instance be also tested") //DEPRECATED
	runsPtr := flag.Int("runs", -1, "how many runs")
	strictPtr := flag.Bool("strict", false, "strict (set) policy for constraints overlap")
	
	flag.Parse()
	
	initVector = InitializationVector{*nPtr, *mPtr, *seedPtr, *consvPtr, *plantedPtr, *runsPtr, *strictPtr} //CONSV IS DEPRECATED
	if initVector.N < 1 || initVector.M < 1 || initVector.Seed < 1 {
		throwError("Flag -n, -m or -seed must be set to run.")
	}
}

func addFootball(startingCosetId int) {

	startingCosetTransitions := cosetTable[startingCosetId].Transitions
	if startingCosetTransitions[0] != 0 || startingCosetTransitions[1] != 0 || startingCosetTransitions[2] != 0 || startingCosetTransitions[3] != 0 {
		panic("can only add a football to a coset with no a/b transitions " + strconv.Itoa(startingCosetId))
	}
	if !cosetTable[startingCosetId].Live {
		panic("cannot add a football to a dead coset")
	}
	lenBefore := len(cosetTable)
	
	for j, cosets := range footballCosets {
		if j==0 {
			for i:=0;i<4;i++ {
				startingCosetTransitions[i] = lenBefore + cosets[i] -2
			}
			continue
		}
		newCosetTransitions := make([]int, (numberOfGenerators + initVector.N)*2)
		for i:=0;i<4;i++ {
			newCosetTransitions[i] = cosets[i] + lenBefore - 2
		}
		newCoset := Coset{lenBefore + j-1,true,newCosetTransitions}
		cosetTable = append(cosetTable, &newCoset)
	}
	cosetTable[lenBefore].Transitions[1] = startingCosetId
	cosetTable[lenBefore+1].Transitions[0] = startingCosetId
	cosetTable[lenBefore+2].Transitions[2] = startingCosetId
	cosetTable[lenBefore+2].Transitions[3] = startingCosetId
}

func pushScan(element ScanCommand) {
	scanQueue = append(scanQueue, element)
}

func scanAll() bool {
	for i:=0;i<len(scanQueue);i++ {
		if scanWord(scanQueue[i]) {
			return true // found proof for unplanted
		}
	}
	scanQueue = []ScanCommand{}
	return false // no proof for unplanted found yet
}

func scanWord(command ScanCommand) bool { // return true, if proof for unplanted instance has been found
	wordLength := len(command.Word)
	startingCoset := command.StartingCoset
	//DEBUG -------------------------------------------------
	if startingCoset > len(cosetTable)-1 {
		fmt.Println("about to fail here:", command)
	}
	//----- -------------------------------------------------
	if(cosetTable[startingCoset].Live) { // else move on
		// scan forward:
		currentCoset := startingCoset
		
		aborted := false
		transversal := make([]int, wordLength)
		for i, gen := range command.Word {
			currentCoset = cosetTable[currentCoset].
				Transitions[getColumnIndex(gen)]
			transversal[i] = currentCoset
			if currentCoset == 0 {
				aborted = true
				break
			}
		}
		
		if aborted { //not the whole word has been scanned
			// scan backwards:
			currentCoset := startingCoset //reset
			for i, _ := range command.Word {
				backwardsGen := (-1)*command.Word[wordLength-1-i]
				nextCoset := cosetTable[currentCoset].Transitions[getColumnIndex(backwardsGen)]
		
				if nextCoset == 0 {
					if i == wordLength-1 { // reached "end" (beginning) of word
						deduction(currentCoset, backwardsGen, startingCoset)
					} else if transversal[wordLength-i-2] != 0 { // deduction
						deduction(currentCoset, backwardsGen, transversal[wordLength-i-2])
						break
					} else { // definition
						defineCoset(currentCoset, backwardsGen)
						currentCoset = len(cosetTable)-1 // this coset has just been defined
					}
				} else {
					currentCoset = nextCoset
					if i == wordLength-1 { // reached "end" (beginning) of word
						if currentCoset != startingCoset {
							// coincidence
							pushCoincidence(startingCoset, currentCoset)
							if processCoincidences() {
								return true // found proof for unplanted
							}
						}
					} else {
						// compare with forwardScan
						if transversal[wordLength-i-2] > 0 && transversal[wordLength-i-2] != currentCoset {
							// coincidence
							pushCoincidence(transversal[wordLength-i-2], currentCoset)
							if processCoincidences() {
								return true // found proof for unplanted
							}
							break	// not sure if we could miss another coincidence, 
								// but there might be a dead coset in the transversal now
						}
					}
				}
			}
		} else if transversal[wordLength-1] != startingCoset {
			// coincidence
			pushCoincidence(transversal[wordLength-1], startingCoset)
			if processCoincidences() {
				return true // found proof for unplanted
			}
		}
	}
	return false // proof for unplanted not yet found, continue
}

func popCoincidence() Coincidence {
	result := coincidenceQueue[0]
	coincidenceQueue = coincidenceQueue[1:]
	return result
}

func pushCoincidence(oneCoset, otherCoset int) {
	if oneCoset > otherCoset { // convention: higher id will survive
		doPushCoincidence(oneCoset, otherCoset)
	} else {
		doPushCoincidence(otherCoset, oneCoset)
	}
}

func doPushCoincidence(higherCoset, lowerCoset int) {
	duplicateFound := false
	for _, coinc := range coincidenceQueue {
		if coinc.Old == higherCoset {
			if sameEqvClass(coinc.New, lowerCoset) {
				duplicateFound = true
				break
			}
		} 
	}
	if !duplicateFound {
		coincidenceQueue = append(coincidenceQueue, Coincidence{higherCoset, lowerCoset})
	}
}

func processCoincidences() bool {
	//init classes
	initEqvClasses()
	
	mergeEqvClasses(coincidenceQueue[0].Old, coincidenceQueue[0].New)
	
	for len(coincidenceQueue) > 0 {
		toProcess := popCoincidence()
		processCoincidence(toProcess)
	}
	
	
	quitEqvClassesMode()

	// ABBRUCHBEDINGUNG ----------------------	
	if cosetTable[1].Transitions[0] == 1 { // found reflexive edge, must be unplanted
		fmt.Println("### Found evidence for UNPLANTED --- Done! ###")
		return true
	}
	// ----------------------------------------

	// NORMALIZING
	//if len(cosetTable) - numberOfLiveCosets() > 1000 {
		//normalizeCosetTable()
	//}
	
	return false
}

func quitEqvClassesMode() {
	// update CosetTable according to Equivalence classes
	for _, coset := range cosetTable {
		if coset.Live {
			for j := 0; j<len(coset.Transitions); j++ {
				cosetRef := coset.Transitions[j]
				if cosetRef > 0 {
					coset.Transitions[j] = eqvClasses[cosetRef].Parent
				}
			}
		}
	}
}

func mergeEqvClasses(oldParent int, newParent int) {
	if oldParent < newParent {
		panic("merging to a class with higher ID")
	}
	childrenToSwitch := eqvClasses[oldParent].Children
	eqvClasses[oldParent] = EqvClass{false, newParent, []int{}}
	eqvClasses[newParent].Children = append(eqvClasses[newParent].Children, oldParent)
	if len(childrenToSwitch) > 0 {
		for _, childId := range childrenToSwitch {
			eqvClasses[childId].Parent = newParent
			eqvClasses[newParent].Children = append(eqvClasses[newParent].Children, childId)
		}
		
	}

}

func sameEqvClass(oneId int, otherId int) bool {
	return eqvClasses[oneId].Parent == eqvClasses[otherId].Parent
}

func initEqvClasses() {
	eqvClasses = make([]EqvClass, len(cosetTable))
	for i:=1; i<len(cosetTable); i++ {
		if cosetTable[i].Live {
			eqvClasses[i] = EqvClass{true, i, []int{}}
		}
	}
}

func processCoincidence(coincidence Coincidence) {
	if coincidence.Old != coincidence.New {
		cosetOld := cosetTable[coincidence.Old]
		cosetNew := cosetTable[coincidence.New]
		cosetOld.Live = false

		// check cosetTable for new coincidences (merge):
		for i, _ := range cosetOld.Transitions {
			if cosetOld.Transitions[i] > 0 {
				if cosetNew.Transitions[i] > 0 {
					transitionNew := cosetNew.Transitions[i]
					transitionOld := cosetOld.Transitions[i]
					if !sameEqvClass(transitionNew, transitionOld) {
						pushCoincidence(transitionOld, transitionNew)

						parentNew := eqvClasses[transitionNew].Parent
						parentOld := eqvClasses[cosetOld.Transitions[i]].Parent

						if parentOld < parentNew {
							mergeEqvClasses(parentNew, parentOld)
						} else {
							mergeEqvClasses(parentOld, parentNew)
						}

						if cosetOld.Transitions[i] < cosetNew.Transitions[i] {
							cosetNew.Transitions[i] = cosetOld.Transitions[i]
						}
					}
				} else {
					cosetNew.Transitions[i] = cosetOld.Transitions[i]
				}
			}
		}
	}
}
func defineCoset(fromCosetId int, generatorTransition int) {
	newCoset := Coset{len(cosetTable), true, make([]int,cosetTableWidth)}
	cosetTable = append(cosetTable, &newCoset)	
	if fromCosetId >= 0 {
		if !cosetTable[fromCosetId].Live {
			fmt.Println("coset id: ", fromCosetId)
			throwError("cannot define a coset transition from a dead coset")
		}
		deduction(fromCosetId, generatorTransition, newCoset.Id)
	}
}

func deduction(fromCosetId int, generatorTransition int, toCosetId int) {
	if generatorTransition == 0 {
		throwError("transition cannot be 0")
	}

	cosetTable[toCosetId].Transitions[getColumnIndex(generatorTransition*(-1))] = fromCosetId
	cosetTable[fromCosetId].Transitions[getColumnIndex(generatorTransition)] = toCosetId
}

func getColumnIndex(generator int) int {
	if generator > 0 {
		return generator*2-2
	} else if generator < 0 {
		return generator*(-2)-1
	} else {
		throwError("generator cannot be 0")
		return 0
	}
}

func printCosetTable(from, until int) {
	// default: print whole table (if until is negative)
	if until < 1 { until = len(cosetTable) }
	if from < 0 { from = 0 }
	
	fmt.Println(" ")
	fmt.Println("Coset table:")
	fmt.Print("     ")
	for i := 1; i <= numberOfGenerators; i++ {
		fmt.Print(i, "    ", i*(-1), "   ")
	}
	fmt.Println()
	for i := from; i<until;i++ {
		c := cosetTable[i]
		if !c.Live { continue } // do not print dead cosets
		if i<10 { fmt.Print(" ") }
		if i<100 { fmt.Print(" ") }
		fmt.Print(c.Id, "; ")
		for i, t := range c.Transitions {
			fmt.Print(t, ";  ")
			if t<10 { fmt.Print(" ") }
			if t<100 { fmt.Print(" ") }
			if i>=18 { fmt.Print(" ") }
		}
		fmt.Println()
	}
	fmt.Println("Number of live cosets: ", numberOfLiveCosets())
}

func numberOfLiveCosets() int {
	result := 0
	for _, coset := range cosetTable {
		if coset.Live { result++ }
	}
	return result
}

func normalizeCosetTable() {
	fmt.Println("Normalizing...")
	removeDeadCosetsAtEndOfTable()
	sort.Ints(deadCosets)
	nrLive := numberOfLiveCosets()
	fmt.Println("length of table is ", len(cosetTable), ", nr of live cosets is ", nrLive)
	for ;nrLive < len(cosetTable) - 1; {
		movingCoset := cosetTable[len(cosetTable)-1]
		
		oldId := movingCoset.Id
		newId := deadCosets[0]
		deadCosets = deadCosets[1:]

		fmt.Println("moving coset ", oldId, " to ", newId)
		movingCoset.Id = newId
		cosetTable[newId] = movingCoset
		for i, tr := range movingCoset.Transitions { // replace Id all over table
			cosetTable[tr].Transitions[i+1-2*(i%2)] = newId // is +-1 away from index in own table
		}
		cosetTable = cosetTable[:oldId-1] // coset has been moved, last line can be deleted
		
		removeDeadCosetsAtEndOfTable()
	}
}

func removeDeadCosetsAtEndOfTable() {
	for ;!cosetTable[len(cosetTable)-1].Live; { // as long as there is a dead coset at the end...
		fmt.Print(cosetTable[len(cosetTable)-1].Id, " ")
		cosetTable = cosetTable[:len(cosetTable)-1] // ...remove it
	}
	fmt.Println(" ")
}

func throwError(message string) {
	panic(message)
}
