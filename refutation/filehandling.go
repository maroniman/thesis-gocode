package main

import  (
	"io/ioutil"
	"strings"
	"strconv"
	"path/filepath"
)

func readFile(absoluteFileName string) ([][]int, [][]int, [][]int, int) {
	contents, err := ioutil.ReadFile(absoluteFileName)
	check(err)
	//parse number of generators (a number in the first line):
	lines := strings.Split(string(contents), "\n")
	numberOfGenerators, err := strconv.Atoi(strings.Trim(lines[0], "\n"))
	check(err)
	lines = lines[1:] //first line is not needed any more
	lines = lines[:len(lines)-1] //last line is end of file
	
	groupRelations := make([][]int, 0, 3)
	constraintRelations := make([][]int, 0, len(lines)-3)
	reverseConstraintRelations := make([][]int, 0, len(lines)-3)
	
	parsingGroupRelations := true //group relations are in first lines, followed by emtpy line
	parsingForwardConstraints := false
	parsingReverseConstraints := false

	for _, line := range lines {
		if line == "" {
			if parsingGroupRelations {
				parsingGroupRelations = false
				parsingForwardConstraints = true
				continue
			} else {
				parsingForwardConstraints = false
				parsingReverseConstraints = true
				continue
			}
		}

		newRelation := parseLine(line)
		if parsingGroupRelations {
			groupRelations = append(groupRelations, newRelation)
		} else if parsingForwardConstraints {
			constraintRelations = append(constraintRelations, newRelation)
		} else if parsingReverseConstraints {
			reverseConstraintRelations = append(reverseConstraintRelations, newRelation)
		} else {
			panic("don't know what to parse")
		}
	}
	return groupRelations, constraintRelations, reverseConstraintRelations, numberOfGenerators
}

func parseLine(line string) []int {
	relation := make([]int, 0, 15)
	generators := strings.Split(line, " ")
	for _, generator := range generators {
		gen, err := strconv.Atoi(generator)	
		check(err)	
		relation = append(relation, gen)
	}
	return relation
}

func getFilename(base string, n int, m int, seed int, planted bool, strictConstraints bool) string {
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
	filenameBuilder.WriteString(".in")
	absPath, error := filepath.Abs(filenameBuilder.String())
	check(error)
	return absPath
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}
