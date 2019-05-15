package main

import (
	"fmt"
	"path/filepath"
	"math/rand"
	"io/ioutil"
	"strings"
	"strconv"
	"flag"
)

type InitializationVector struct {
	N int
	M int
	Seed int
	Runs int
}

var initVector InitializationVector = InitializationVector{-1,-1, 1, 1}

const arity int = 3
const nrA5Generators int = 2

var isEliminated []bool

func main() {
	initializeInitVector()
	fmt.Println("Read(\"path-to-here/simplify-constraints/instances/gap-BEFORE-a5_0n"+strconv.Itoa(initVector.N)+"m"+ strconv.Itoa(initVector.M)+"-"+strconv.Itoa(initVector.Seed)+"\");;runTook;")
	rand.Seed(int64(initVector.Seed))
	foundProof := 0
	for run:=0; run<initVector.Runs; run++ {
		constraints := generateConstraints(initVector.N, initVector.M)
		constraints = appendConstants(constraints)
		
		isEliminated = make([]bool, initVector.N+1+nrA5Generators)
		
		filename := getFilename("instances/gap-BEFORE-a5_"+strconv.Itoa(run), initVector.N, initVector.M, initVector.Seed)
		writeInstanceFiles(constraints, filename)
		
		done := false
		unplantedProof := false
		for ;!done; {
			i,j,k,l := tryPairs(constraints)
			if i < 0 {
				done = true
			} else {
				//fmt.Println("pair:", constraints[i], constraints[k])
				constraints, unplantedProof = resolveEquivalence(i,j,k,l, constraints)
				if unplantedProof {
					done = true
					//fmt.Print("--")
					foundProof++
				} 
			}
		}

		//fmt.Println("el", initVector.M-len(constraints), "(", run, ")")
		if !unplantedProof {
			filename := getFilename("instances/gap-SIMPLIFIED-a5_"+strconv.Itoa(run), initVector.N, initVector.M, initVector.Seed)
			writeInstanceFiles(constraints, filename)
		}
	}
	fmt.Println(foundProof,"/",initVector.Runs)
}

func initializeInitVector() {
	nPtr := flag.Int("n", -1, "number of variables")
	mPtr := flag.Int("m", -1, "number of constraints")
	seedPtr := flag.Int("seed", -1, "initial seed")
	runsPtr := flag.Int("runs", 1, "number of runs")
	
	flag.Parse()
	initVector = InitializationVector{*nPtr, *mPtr, *seedPtr, *runsPtr}
	if initVector.N < 1 || initVector.M < 1 || initVector.Seed < 1 {
		panic("Flag -n, -m or -seed must be set to run.")
	}
}

func resolveEquivalence(i,j,k,l int, constraints [][]int) ([][]int, bool) {
	ind1 := findThirdVariable(constraints[i], j)
	ind2 := findThirdVariable(constraints[k], l)
	
	var higherConstraint int
	if constraints[i][ind1] < constraints[k][ind2] {
		higherConstraint = k
		replace(constraints[i], j, ind1, constraints[k], l, ind2, constraints)
	} else if constraints[k][ind2] < constraints[i][ind1] {
		higherConstraint = i
		replace(constraints[k], l, ind2, constraints[i], j, ind1, constraints)
	} else {
		return constraints, true
	}

	return deleteConstraint(higherConstraint, constraints), false
}

func replace(lowerCons []int, lowerPairInd int, lowerInd int, higherCons []int, higherPairInd int, higherInd int, constraints [][]int) {
	isEliminated[higherCons[higherInd]] = true
	
	//fmt.Println(higherCons[higherInd], "-->", lowerCons[lowerInd])
	lowerRest := getRestOfConstraint(lowerCons, lowerPairInd)
	higherRest := getRestOfConstraint(higherCons, higherPairInd)	

	//fmt.Println("rests", lowerCons, lowerRest)
	//fmt.Println("rests", higherCons, higherRest)

	replacer := []int{}

	index := findIndex(higherRest, higherCons[higherInd])
	if index > 0 {
		replacer = append(replacer, reverseWord(higherRest[:index])...)
	}
	replacer = append(replacer, lowerRest...)
	if index < len(higherRest)-1 {
		replacer = append(replacer, reverseWord(higherRest[index+1:])...)
	}

	replaceVariable(higherCons[higherInd], replacer, constraints)
}


func getRestOfConstraint(cons []int, pairStarting int) []int {
	rest := make([]int, len(cons)-2)
	if pairStarting==len(cons)-1 {
		copy(rest, cons[1:len(cons)-1])
	} else {
		if pairStarting!=len(cons)-2 {
			copy(rest[:len(cons)-pairStarting-2],cons[pairStarting+2:])
		}
		if pairStarting!=0 {
			copy(rest[len(cons)-pairStarting-2:], cons[:pairStarting])
		}
	}
	return rest
}

func findThirdVariable(cons []int, pairStarting int) int {
	for i := 0; i<len(cons)-2; i++ {
		ind := (pairStarting + 2 + i)%len(cons)
		if isVariable(cons[ind]) {
			return ind
		}
	}
	return -1
}

func replaceVariable(oldVar int, replacer []int, constraints [][]int) {
	//fmt.Println("replacing", oldVar, "by", replacer)
	for i:=0;i<len(constraints);i++ {
		for j:=0; j<len(constraints[i]); j++ {
			if constraints[i][j] == oldVar {
				newConstraint := make([]int, len(constraints[i])+len(replacer)-1)
				if j!=0 {
					copy(newConstraint[:j], constraints[i][:j])
				}
				copy(newConstraint[j:len(replacer)+j], replacer[:])
				if j!=len(constraints[i])-1 {
					copy(newConstraint[len(replacer)+j:], constraints[i][j+1:])
				}
				//fmt.Println("was:", constraints[i], "-> is:", newConstraint)
				constraints[i] = newConstraint
				break
			}
		}
	}
}

func deleteConstraint(line int, constraints [][]int) [][]int {
	newConstraints := make([][]int, len(constraints)-1)
	copy(newConstraints[:line], constraints[:line])
	if line < len(constraints) -1 {
		copy(newConstraints[line:], constraints[line+1:])
	}
	return newConstraints
}

func tryPairs(constraints [][]int) (int,int,int,int){
	for i := 0; i<len(constraints); i++ {
		for j := 0; j<len(constraints[i]); j++ {
			if !isVariable(constraints[i][j]) || !isVariable(constraints[i][(j+1)%len(constraints[i])]) {
				continue
			}
			for k := i+1; k<len(constraints); k++ {
				for l := 0; l<len(constraints[k]); l++ {
					if !isVariable(constraints[k][l]) || !isVariable(constraints[k][(l+1)%len(constraints[k])]) {
						continue
					}
					if constraints[i][j] == constraints[k][l] && constraints[i][(j+1)%len(constraints[i])] == constraints[k][(l+1)%len(constraints[k])] {
						//fmt.Println("pair:", i, j, k, l)
						return i, j, k, l
					}
				}
			}
		}
	}
	return -1,-1,-1,-1
}

func isVariable(i int) bool {
	return i > nrA5Generators
}

func generateConstraints(n int, m int) [][]int {
	constraints := make([][]int, 0, m)
	for i:= 0; i<m; i++ {
		currentConstraint := make([]int, 0, arity)
		
		for j := 0; j<arity; j++ {
			newVariable := 0
			for redrawn:=true; redrawn; {
				newVariable = nrA5Generators + 1 + rand.Intn(n)
				redrawn = isRedraw(newVariable, currentConstraint)
			}
			currentConstraint = append(currentConstraint, newVariable)
		}
		//fmt.Println(i,":", currentConstraint)
		constraints = append(constraints, currentConstraint)
	}

	return constraints
}

func isRedraw(newVariable int, constraint []int) bool {
	for i:=0; i<len(constraint); i++ {
		if constraint[i] == newVariable {
			return true
		}
	}
	return false //return immediately if constraint is empty
}

func appendConstants(constraints [][]int) [][]int {
	newConstraints := make([][]int, len(constraints))
	for i := 0; i<len(constraints); i++ {
		newConstraints[i] = append(constraints[i], reverseWord(wordFromPermutation(randomEven5Permutation()))...)
	}
	return newConstraints
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

func writeInstanceFiles(constraints [][]int, fileName string) {
	var constraintBuilder strings.Builder

	constraintBuilder.WriteString("CosetTableDefaultMaxLimit := 131072000;;\nstartTime := Runtime();;\nf:=FreeGroup(\"x1\", \"x2\"") 
	
	for i:= 1+nrA5Generators; i<initVector.N+1+nrA5Generators; i++ {
		if !isEliminated[i] {
			constraintBuilder.WriteString(", \"x")
			constraintBuilder.WriteString(strconv.Itoa(getRealIndex(i)))
			constraintBuilder.WriteString("\"")
		}
	}
	
	constraintBuilder.WriteString(");; g:= f/[f.1^5, f.2^2, (f.1*f.2)^3")

	for i,cons := range constraints {
		constraintBuilder.WriteString(",\n")
		for j:=0;j<len(cons);j++ {
			constraintBuilder.WriteString("f.")
			constraintBuilder.WriteString(strconv.Itoa(getRealIndex(constraints[i][j])))
			if constraints[i][j] < 0 {
				constraintBuilder.WriteString("^-1")	
			} 
			
			if !(j==len(cons)-1) { // only if this is not identity
				constraintBuilder.WriteString("*")
			}
		}
	}
	
	constraintBuilder.WriteString("];;\ntab := CosetTable(g, Subgroup(g, []));;\nrunTook := StringTime(Runtime()-startTime);;\nlen:=Length(TransposedMat(tab));;")
	
	err := ioutil.WriteFile(fileName, []byte(constraintBuilder.String()), 0644)
	check(err)
}

func getRealIndex(oldIndex int) int {
	result := oldIndex
	if result < 0 {
		result = result*-1
	}
	for i:=1;i<oldIndex;i++ {
		if isEliminated[i] {
			result--
		}
	}
	return result
}

func getFilename(base string, n int, m int, seed int) string {
	var filenameBuilder strings.Builder
	filenameBuilder.WriteString(base)
	filenameBuilder.WriteString("n")
	filenameBuilder.WriteString(strconv.Itoa(n))
	filenameBuilder.WriteString("m")
	filenameBuilder.WriteString(strconv.Itoa(m))
	filenameBuilder.WriteString("-")
	filenameBuilder.WriteString(strconv.Itoa(seed))
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
