package main

import (
	"fmt"
	"flag"
	"path/filepath"
	"math/rand"
	"io/ioutil"
	"strings"
	"strconv"
)

type InitializationVector struct {
	N int
	M int
	Seed int
	StrictlyDifferentConstraints bool
	X60 bool
}

var initVector InitializationVector = InitializationVector{-1,-1, 1, false, false}

const arity int = 3
const nrA5Generators int = 2

func main() {
	fmt.Println("started. GAP command to copy:")
	initializeInitVector()
	generateInstance()
	
	strictString := ""
	if initVector.StrictlyDifferentConstraints {
		strictString = "_s"
	}

	fmt.Println("Read(\"path-to-here/generate-gap-instances/instances/gap-a5n"+strconv.Itoa(initVector.N)+"m"+ strconv.Itoa(initVector.M)+"-"+strconv.Itoa(initVector.Seed)+"_u"+strictString+"\");;runTook;"+"Read(\"path-to-here/generate-gap-instances/instances/gap-a5n"+strconv.Itoa(initVector.N)+"m"+ strconv.Itoa(initVector.M)+"-"+strconv.Itoa(initVector.Seed)+strictString+"\");;runTook;")
}

func initializeInitVector() () {
	nPtr := flag.Int("n", -1, "number of variables")
	mPtr := flag.Int("m", -1, "number of constraints")
	seedPtr := flag.Int("seed", -1, "initial seed")
	strictPtr := flag.Bool("strict", false, "strict (set) policy for constraints overlap")
	x60Ptr := flag.Bool("x60", false, "add x^60 constraints")
	
	flag.Parse()
	initVector = InitializationVector{*nPtr, *mPtr, *seedPtr, *strictPtr, *x60Ptr}
	if initVector.N < 1 || initVector.M < 1 || initVector.Seed < 1 {
		panic("Flag -n, -m or -seed must be set to run.")
	}
}

func generateInstance() {
	rand.Seed(int64(initVector.Seed))
	constraints := generateConstraints(initVector.N, initVector.M, initVector.StrictlyDifferentConstraints)
	assignment := generateA5Assignment(initVector.N)
	reverseRightHalf := computeReverseRightHalf(constraints, assignment)
	randomRightHalf := generateRandomRightHalf(initVector.M)

	plantedName := getFilename("instances/gap-a5", initVector.N, initVector.M, initVector.Seed, true, initVector.StrictlyDifferentConstraints, initVector.X60)
	uplantedName := getFilename("instances/gap-a5", initVector.N, initVector.M, initVector.Seed, false, initVector.StrictlyDifferentConstraints, initVector.X60)

	writeInstanceFiles(constraints, reverseRightHalf, randomRightHalf, initVector.N, plantedName, uplantedName)
}

func generateRandomRightHalf(m int) [][]int {
	result := make([][]int, m)
	for i, _ := range result {
		result[i] = reverseWord(wordFromPermutation(randomEven5Permutation()))
	}
	return result
}

func computeReverseRightHalf(constraints [][arity]int, assignment [][]int) [][]int {
	m := len(constraints)	
	reverseRightHalf := make([][]int, m)
	
	for i := 0; i < m; i++ {
		perm := []int{0,1,2,3,4}
		for j:=0;j<arity;j++ {
			perm = applyPermutation(perm, assignment[constraints[i][j]])
		}
		//fmt.Println("right half perm", i, ": ", perm)
		reverseRightHalf[i] = reverseWord(wordFromPermutation(perm))
		//fmt.Println("right half ", i, ": ", reverseRightHalf[i])
	}
	
	return reverseRightHalf
}

func wordFromPermutation(originalPerm []int) []int {
	wordx := []int{-1, 2, 1, 1} //Abaa
	wordX := []int{-1, -1, -2, 1} //AABa

	a := []int{4, 0, 1, 2, 3} // generator 1
	b := []int{1, 0, 3, 2, 4} // generator 2
	X := []int{1, 3, 2, 0, 4}
	x := []int{3, 0, 2, 1, 4}

	result := make([]int, 0)
	perm := make([]int, len(originalPerm))
	copy(perm, originalPerm)

	for ;findIndex(perm,4)<4; {
		perm = applyPermutation(perm, a)
		result = append(result, 1) // a
	}
	if findIndex(perm,2)!=2 {
		switch findIndex(perm,2) {
			case 0: // first X=xx...
				perm = applyPermutation(perm, X)
				result = append(result, wordX...)
			case 1: // first x...
				perm = applyPermutation(perm, x)
				result = append(result, wordx...)
			// case 3: only b is needed (see next line)
		}
		//swap once (operation b) to get the 2 into place:
		perm = applyPermutation(perm, b)
		result = append(result, 2)
	}
	switch findIndex(perm,0) {
			// case 0: nothing to do
			case 1: //X
				perm = applyPermutation(perm, X)
				result = append(result, wordX...)
			case 3: //x 
				perm = applyPermutation(perm, x)
				result = append(result, wordx...)
	}

	return reverseWord(result) // reverse because we have reverse engineered the word
}

func reverseWord(originalWord []int) []int {
	word := make([]int, len(originalWord))
	copy(word, originalWord)
	result := make([]int, 0, len(word))
	for ;len(word)>0; {
		result = append(result, word[len(word)-1]*(-1)) //append reverse generator in reverse order
		word = word[:len(word)-1] //remove last generator from wordcopy
	}
	return result
}

func applyPermutation(toPermute, perm []int) []int {
	result := make([]int, 5)
	for i, _ := range toPermute {
		result[i] = toPermute[perm[i]]
	}
	return result
}

func generateA5Assignment(n int) [][]int {
	assignment := make([][]int, n+1+nrA5Generators)	
	for i := 1+nrA5Generators; i<n+1+nrA5Generators; i++ {
		assignment[i] = randomEven5Permutation()
		//fmt.Println("assignment ", i, ": ", assignment[i])
	}
	return assignment
}

func randomEven5Permutation() []int {
	result := rand.Perm(5)
	for ;!isEvenPermutation(result); {
		result = rand.Perm(5)
	}
	return result
}

func isEvenPermutation(permToTest []int) bool {
	perm := make([]int, 5)
	copy(perm, permToTest)
	swaps := 0
	for i := 0; i<4; i++ {
		index := findIndex(perm, i)
		// element i is at perm[index]
		for j:=index; i<j; j-- {
			//swap perm[j] and perm[j-1]
			perm[j] = perm[j-1]
			perm[j-1] = i
			swaps ++
		}
	}	
	return swaps%2 == 0
}

func findIndex(arr []int, el int) int {
	index := -1
	for i,_ := range arr {
			if arr[i] == el {
				index = i
			}
		}
	return index
}

func generateConstraints(n int, m int, strictlyDifferentConstraints bool) [][arity]int {
	constraints := make([][arity]int, 0, m)
	for i:= 0; i<m; i++ {
		var currentConstraint [arity]int
		
		for j := 1; j<arity; j++ {
			if j == 1 {
				currentConstraint[0] = nrA5Generators + 1 + rand.Intn(n) // no duplicate checking needed for first part, so it is generated here
			}
			currentConstraint[j] = nrA5Generators + 1 + rand.Intn(n)
			
			acceptable := false // depends on policy
			if strictlyDifferentConstraints {
				acceptable = isAllDifferent(currentConstraint, j) && !isDuplicateStrict(constraints, currentConstraint, j)
			} else {
				acceptable = isAllDifferent(currentConstraint, j) && !isDuplicateWeak(constraints, currentConstraint, j)
			}
			
			if (!acceptable) {
				j = 0 //will be incremented to 1, started from beginning
			}
		}
		//fmt.Println("New constraint: ", currentConstraint)
		constraints = append(constraints, currentConstraint)
	}
	return constraints
}

func isAllDifferent(constraintToCheck [arity]int, upToIndex int) bool {
	for i := 0; i<=upToIndex; i++ {
		for j := 0; j<=upToIndex; j++ {
			if j!=i && constraintToCheck[i] == constraintToCheck[j] {
				return false
			}
		}
	}
	return true
}

func isDuplicateStrict(constraints [][arity]int, constraintToCheck [arity]int, upToIndex int) bool { 
// Set notion: no more than 1 variable overlap between two constraints
	isDuplicate := false // in case this is first constraint, immediate return
	for _, cons := range constraints {
		overlap := 0
		for i := 0; !(overlap > 1) && i<arity; i++ {
			for j := 0; j<=upToIndex; j++ {
				if cons[i] == constraintToCheck[j] {
					overlap++
				}
			}
		}
		if overlap > 1 { 
			isDuplicate = true
			break
		}
	}
	return isDuplicate
}

func isDuplicateWeak(constraints [][arity]int, constraintToCheck [arity]int, upToIndex int) bool { // "weaker" standard notion: pairwise check
	isDuplicate := false
	for _, cons := range constraints {
		isDuplicate = true
		for i := 0; isDuplicate && i<=upToIndex; i++ {
			isDuplicate = cons[i] == constraintToCheck[i]
		}
		if isDuplicate { 
			break
		}
		if upToIndex > 1 {
			isDuplicate = true
			for i := upToIndex; isDuplicate && i>=1; i-- {
				isDuplicate = cons[i] == constraintToCheck[i]
			}
			if isDuplicate { 
				break
			}
		}
	}
	return isDuplicate
}

func writeInstanceFiles(constraints [][arity]int, reverseRightHalf [][]int, randomRightHalf [][]int, n int, plantedFileName string, unplantedFileName string) {
	var constraintBuilder strings.Builder
	var unplantedBuilder strings.Builder

	constraintBuilder.WriteString("CosetTableDefaultMaxLimit := 131072000;;\nstartTime := Runtime();;\nf:=FreeGroup(\"x1\", \"x2\"") 
	
	for i:= 1+nrA5Generators; i<n+1+nrA5Generators; i++ {
		constraintBuilder.WriteString(", \"x")
		constraintBuilder.WriteString(strconv.Itoa(i))
		constraintBuilder.WriteString("\"")
	}
	
	constraintBuilder.WriteString(");; g:= f/[f.1^5, f.2^2, (f.1*f.2)^3")

	unplantedBuilder.WriteString(constraintBuilder.String()) // same constraint but now it will get different

	for i,_ := range constraints {
		constraintBuilder.WriteString(",\n")
		unplantedBuilder.WriteString(",\n")
		for j:=0;j<arity;j++ {
			constraintBuilder.WriteString("f.")
			unplantedBuilder.WriteString("f.")
			constraintBuilder.WriteString(strconv.Itoa(constraints[i][j]))
			unplantedBuilder.WriteString(strconv.Itoa(constraints[i][j]))
			
			if !(j==arity-1 && len(reverseRightHalf[i]) == 0) { // only if this is not identity
				constraintBuilder.WriteString("*")
			}
			if !(j==arity-1 && len(randomRightHalf[i]) == 0) { // only if this is not identity
				unplantedBuilder.WriteString("*")
			}
		}
		
		for k,_ := range reverseRightHalf[i] {
			constraintBuilder.WriteString("f.")
			if reverseRightHalf[i][k] > 0 {
				constraintBuilder.WriteString(strconv.Itoa(reverseRightHalf[i][k]))
			} else {
				constraintBuilder.WriteString(strconv.Itoa((-1)*reverseRightHalf[i][k]))
				constraintBuilder.WriteString("^-1")
			}
			
			if k<len(reverseRightHalf[i])-1 { //no * after last
				constraintBuilder.WriteString("*")
			}
		}
		
		for k,_ := range randomRightHalf[i] {
			unplantedBuilder.WriteString("f.")
			if randomRightHalf[i][k] > 0 {
				unplantedBuilder.WriteString(strconv.Itoa(randomRightHalf[i][k]))
			} else {
				unplantedBuilder.WriteString(strconv.Itoa((-1)*randomRightHalf[i][k]))
				unplantedBuilder.WriteString("^-1")
			}
			
			if k<len(randomRightHalf[i])-1 { //no * after last
				unplantedBuilder.WriteString("*")
			}
		}	
	}

	if initVector.X60 {
		for j:= 1+nrA5Generators; j<n+1+nrA5Generators; j++ {
			constraintBuilder.WriteString(",\n")
			constraintBuilder.WriteString("f.")
			constraintBuilder.WriteString(strconv.Itoa(j))
			constraintBuilder.WriteString("^60")

			unplantedBuilder.WriteString(",\n")
			unplantedBuilder.WriteString("f.")
			unplantedBuilder.WriteString(strconv.Itoa(j))
			unplantedBuilder.WriteString("^60")

			for k:= 1+nrA5Generators; k<n+1+nrA5Generators; k++ {
				if k != j {
					constraintBuilder.WriteString(",\n(")
					constraintBuilder.WriteString("f.")
					constraintBuilder.WriteString(strconv.Itoa(k))
					constraintBuilder.WriteString("*")
					constraintBuilder.WriteString("f.")
					constraintBuilder.WriteString(strconv.Itoa(j))
					constraintBuilder.WriteString(")^60")
					
					unplantedBuilder.WriteString(",\n(")
					unplantedBuilder.WriteString("f.")
					unplantedBuilder.WriteString(strconv.Itoa(k))
					unplantedBuilder.WriteString("*")
					unplantedBuilder.WriteString("f.")
					unplantedBuilder.WriteString(strconv.Itoa(j))
					unplantedBuilder.WriteString(")^60")
					for l:= 1+nrA5Generators; l<n+1+nrA5Generators; l++ {
						if l != k && l != j {
							constraintBuilder.WriteString(",\n(")
							constraintBuilder.WriteString("f.")
							constraintBuilder.WriteString(strconv.Itoa(k))
							constraintBuilder.WriteString("*")
							constraintBuilder.WriteString("f.")
							constraintBuilder.WriteString(strconv.Itoa(j))
							constraintBuilder.WriteString("*")
							constraintBuilder.WriteString("f.")
							constraintBuilder.WriteString(strconv.Itoa(l))
							constraintBuilder.WriteString(")^60")
					
							unplantedBuilder.WriteString(",\n(")
							unplantedBuilder.WriteString("f.")
							unplantedBuilder.WriteString(strconv.Itoa(k))
							unplantedBuilder.WriteString("*")
							unplantedBuilder.WriteString("f.")
							unplantedBuilder.WriteString(strconv.Itoa(j))
							unplantedBuilder.WriteString("*")
							unplantedBuilder.WriteString("f.")
							unplantedBuilder.WriteString(strconv.Itoa(l))
							unplantedBuilder.WriteString(")^60")
						}
					}
				}
			}
		}
	}
	
	unplantedBuilder.WriteString("];;\ntab := CosetTable(g, Subgroup(g, []));;\nrunTook := StringTime(Runtime()-startTime);;\nlenu:=Length(TransposedMat(tab));;")
	constraintBuilder.WriteString("];;\ntab := CosetTable(g, Subgroup(g, []));;\nrunTook := StringTime(Runtime()-startTime);;\nlenp:=Length(TransposedMat(tab));;")
	
	err := ioutil.WriteFile(plantedFileName, []byte(constraintBuilder.String()), 0644)
	check(err)
	err = ioutil.WriteFile(unplantedFileName, []byte(unplantedBuilder.String()), 0644)
	check(err)
}

func getFilename(base string, n int, m int, seed int, planted bool, strictConstraints bool, x60 bool) string {
	var filenameBuilder strings.Builder
	filenameBuilder.WriteString(base)
	filenameBuilder.WriteString("n")
	filenameBuilder.WriteString(strconv.Itoa(n))
	filenameBuilder.WriteString("m")
	filenameBuilder.WriteString(strconv.Itoa(m))
	filenameBuilder.WriteString("-")
	filenameBuilder.WriteString(strconv.Itoa(seed))
	if !planted {
		filenameBuilder.WriteString("_u")
	}
	if strictConstraints {
		filenameBuilder.WriteString("_s")
	}
	if x60 {
		filenameBuilder.WriteString("_x")
	}
	//filenameBuilder.WriteString(".txt")
	absPath, error := filepath.Abs(filenameBuilder.String())
	check(error)
	return absPath
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}
