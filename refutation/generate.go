package main

import (
	"fmt"
	"math/rand"
	"io/ioutil"
	"strings"
	"strconv"
)

const arity int = 3
const nrA5Generators int = 2

func generateInstance(n int, m int, seed int64, strictlyDifferentConstraints bool) {
	rand.Seed(seed)
	constraints := generateConstraints(n, m, seed, strictlyDifferentConstraints)
	assignment := generateA5Assignment(n)
	reverseRightHalf := computeReverseRightHalf(constraints, assignment)
	randomRightHalf := generateRandomRightHalf(m)

	writeInstanceFiles(constraints, reverseRightHalf, randomRightHalf, n, getFilename("../data/a5", n, m, int(seed), true, strictlyDifferentConstraints), getFilename("../data/a5", n, m, int(seed), false, strictlyDifferentConstraints))
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
	fmt.Print("  variables with id assignment: ")
	for i := 1+nrA5Generators; i<n+1+nrA5Generators; i++ {
		assignment[i] = randomEven5Permutation()
		//comparison
		same := true
		for k :=0; k<5;k++ {
			same = same && assignment[i][k] == k
		}
		
		if same {
			fmt.Print(i)
			fmt.Print(", ")
		}
		//fmt.Println("assignment ", i, ": ", assignment[i])
	}
	fmt.Println(".")
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

func generateConstraints(n int, m int, seed int64, strictlyDifferentConstraints bool) [][arity]int {
	constraints := make([][arity]int, 0, m)
	var pairs strings.Builder // only for non-strict: remember pairs which might help in solving
	for i:= 0; i<m; i++ {
		pairSaved := false // flag if duplicating "pair" was remembered in case the constraint is re-drawn
		var currentConstraint [arity]int
		
		for j := 1; j<arity; j++ {
			if j == 1 {
				currentConstraint[0] = nrA5Generators + 1 + rand.Intn(n) // no duplicate checking needed for first part, so it is generated here
			}
			currentConstraint[j] = nrA5Generators + 1 + rand.Intn(n)
			
			acceptable := false // depends on policy
			if strictlyDifferentConstraints {
				acceptable = isAllDifferent(currentConstraint, j) && findDuplicateStrict(constraints, currentConstraint, j) == -1
			} else {
				acceptable = isAllDifferent(currentConstraint, j) && !isDuplicateWeak(constraints, currentConstraint, j)
				dupl := findDuplicateStrict(constraints, currentConstraint, j)
				if acceptable && dupl > 0 { // save "pair" which might help in solving
					pairs.WriteString(strconv.Itoa(i) + ": " + strconv.Itoa(dupl)+"\n")
					pairSaved = true
				}
			}
			
			if (!acceptable) {
				j = 0 //will be incremented to 1, started from beginning
				if pairSaved {
					pairs.WriteString(strconv.Itoa(i) + "redrawn.\n")
					pairSaved = false
				}
			}
		}
		//fmt.Println("New constraint: ", currentConstraint)
		constraints = append(constraints, currentConstraint)
	}

	if len(pairs.String()) > 0 { //write pairs to file
		filename := getFilename("../data/pairs/pairs-a5", n, m, int(seed), true, false)
		writePairsFile(pairs.String(), filename)
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

func findDuplicateStrict(constraints [][arity]int, constraintToCheck [arity]int, upToIndex int) int { // returns index of first constraint with overlap >1
// Set notion: no more than 1 variable overlap between two constraints
	duplicateConstraint := -1 // in case this is first constraint, immediate return
	for consIndex, cons := range constraints {
		overlap := 0
		for i := 0; !(overlap > 1) && i<arity; i++ {
			for j := 0; j<=upToIndex; j++ {
				if cons[i] == constraintToCheck[j] {
					overlap++
				}
			}
		}
		if overlap > 1 { 
			duplicateConstraint = consIndex
			break
		}
	}
	return duplicateConstraint
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

func writePairsFile(pairs, filename string) {
	err := ioutil.WriteFile(filename, []byte(pairs), 0644)
	check(err)
}

func writeInstanceFiles(constraints [][arity]int, reverseRightHalf [][]int, randomRightHalf [][]int, n int, plantedFileName string, unplantedFileName string) {
	//m := len(constraints)
	var constraintBuilder strings.Builder
	var unplantedBuilder strings.Builder

	constraintBuilder.WriteString(strconv.Itoa(n+nrA5Generators))
	constraintBuilder.WriteString("\n1 1 1 1 1\n2 2\n1 2 1 2 1 2\n\n")

	unplantedBuilder.WriteString(constraintBuilder.String()) // same constraint but now it will get different

	for i,_ := range constraints {
		for j:=0;j<arity;j++ {
			constraintBuilder.WriteString(strconv.Itoa(constraints[i][j]))
			unplantedBuilder.WriteString(strconv.Itoa(constraints[i][j]))
			if arity-j>1 { //no space after last
				constraintBuilder.WriteString(" ")
				unplantedBuilder.WriteString(" ")
			}
		}
		
		for k,_ := range reverseRightHalf[i] {
			constraintBuilder.WriteString(" " + strconv.Itoa(reverseRightHalf[i][k]))
		}
		for k,_ := range randomRightHalf[i] {
			unplantedBuilder.WriteString(" " + strconv.Itoa(randomRightHalf[i][k]))
		}
		unplantedBuilder.WriteString("\n")
		constraintBuilder.WriteString("\n")
	}
	
	// REVERSE CONSTRAINTS: First abaababa... then the constraint variables
	constraintBuilder.WriteString("\n")
	unplantedBuilder.WriteString("\n")
	for i,_ := range constraints {		
		for k,_ := range reverseRightHalf[i] {
			constraintBuilder.WriteString(strconv.Itoa(reverseRightHalf[i][k]) + " ")
		}
		for k,_ := range randomRightHalf[i] {
			unplantedBuilder.WriteString(strconv.Itoa(randomRightHalf[i][k]) + " ")
		}
		for j:=0;j<arity;j++ {
			if j>0 { //no space before first
				constraintBuilder.WriteString(" ")
				unplantedBuilder.WriteString(" ")
			}
			constraintBuilder.WriteString(strconv.Itoa(constraints[i][j]))
			unplantedBuilder.WriteString(strconv.Itoa(constraints[i][j]))
		}
		constraintBuilder.WriteString("\n")
		unplantedBuilder.WriteString("\n")
		
	}
	
	err := ioutil.WriteFile(plantedFileName, []byte(constraintBuilder.String()), 0644)
	check(err)
	err = ioutil.WriteFile(unplantedFileName, []byte(unplantedBuilder.String()), 0644)
	check(err)
}
