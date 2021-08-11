package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//we don't care about the name, just a unique reference
type ContigMap struct {
	contigNum map[string]int
	index int
}

func (m *ContigMap) Put(key string) int {
	//check if the key exists first
	index, ok := m.contigNum[key]
	if ok {
		return index
	} else {
		//if it doesn't exist, update the index, update the maps and return the new index
		m.index++
		m.contigNum[key] = m.index
		return m.index
	}
}


//sample name with associated rows
type Sample struct {
	*sync.Mutex
	name string
	contigs map[int]Rows
}


//single row of input
type Row struct {
	pos uint32
	base byte
}

type Rows []Row

var countNonBases = flag.Bool("all", false, "include all characters in the base column, not just bases")
var numSamples = flag.Int("samples", 0, "number of unique sample ids")

func validBase(base byte) bool {
	if *countNonBases {
		return true
	}
	//check if valid base
	switch base {
	case 65: //A
		return true
	case 84: //T
		return true
	case 67: //C
		return true
	case 71: //G
		return true
	default: //none ofthe above
		return false
	}
}

func main() {
	//handle the flags
	flag.Parse()

	if *numSamples < 2 {
		log.Fatal("Must supply number of samples > 2 with -samples X")
	}
	
	//make the results early (so we can start goroutines asap)
	result := make([][]int, *numSamples)
	for i:=0; i<*numSamples; i++ {
		result[i] = make([]int, *numSamples)
	}
	
	samples := []Sample{}
	//discard the header
	read := bufio.NewReader(os.Stdin)
	r := csv.NewReader(read)
	_, err := r.Read()
	if err != nil {
		log.Fatal("dodgy header row read")
	}

	contigMap := ContigMap{
		contigNum: make(map[string]int),
	}

	row, _ := r.Read()

	post, err := strconv.ParseUint(row[2], 10, 32)
	if err != nil {
		log.Fatal("couldn't convert string to uint64")
	}
	pos := uint32(post)
	contigId := contigMap.Put(row[1])
	base := row[3][0]

	firstRow := Row{
		// contigRef: uint8(contigId),
		pos: pos,
		base: base,
	}

	//we keep a memory a memory of the sample, so we can add rows, then append to []samples when filled
	//each complete sample is inserted into samples 

	memSample := Sample{
		name: row[0],
		// rows: []Row{firstRow},
		contigs: map[int]Rows{
			contigId: []Row{
				firstRow,
			},
		},
	}

	var wg sync.WaitGroup
	sampleIndex := 0 //keep track of which sample group we're up to
	for {
		row, err := r.Read()
		if err != nil {
			//this should be the last sample, save
			samples = append(samples, memSample)
			break
		}

		post, err := strconv.ParseUint(row[2], 10, 32)
		if err != nil {
			log.Fatal("couldn't convert string to uint64")
		}
		pos := uint32(post)
		contigId := contigMap.Put(row[1])
		base := row[3][0]

		newRow := Row{
			pos: pos,
			base: base,
		}


		if memSample.name != row[0] {
			//new row, so commiit the existing memSample to the slice
			samples = append(samples, memSample)
			//create a new one
			memSample = Sample{
				name: row[0],
				contigs: map[int]Rows{
					contigId: []Row{
						newRow,
					},
				},
			}
			sampleIndex++

		} else {
			memSample.contigs[contigId] = append(memSample.contigs[contigId], newRow)
		}
	}

	//we now have an array of rows, with a mapping to the corresponding contig and sample name
	// wg.Wait()


	
	if *countNonBases{
		//find the non-base positions and remove all associated rows
		nonBasePositions := collectSortedNonBasePositions(samples)
		var samWg sync.WaitGroup
		for _, sample := range samples {
			samWg.Add(1)
			go sample.removeRowsOfPos(nonBasePositions, &samWg)
		}
		samWg.Wait()
	}

	for i:=1; i<len(samples); i++ {
		//count up to but not including TODO as another iter
		//eachs hould compare to all others, not back to self, not self

		wg.Add(1)
		// go setResults(i, sampleIndex, &wg, &samples[i], &samples[sampleIndex], result)
	}
	wg.Wait()

	outputResults(result, samples)
	
}

// //goes through rows until there's a valid one
// func nextValidRow(r *csv.Reader) ([]string, error) {
// 	row, err := r.Read()
// 	if err != nil {
// 		return row, err
// 	}
// 	if validBase(row[3][0]) {
// 		return row, err
// 	} else {
// 		return nextValidRow(r)
// 	}
// }


//takes the index of two sample, along with the two samples to compare, sets the output in the results matrix
func setResults(i1 int, i2 int, wg *sync.WaitGroup, s1 *Sample, s2 *Sample, result [][]int) {
	defer wg.Done()
	//smallest takes left position
	diffs := s1.contigDiffs(s2)

	//save results in chiral matrix table (both places)
	result[i1][i2] = diffs
	result[i2][i1] = diffs
}

//handle getting the sum of diffs between the two grouped contigs
func (sam1 *Sample)contigDiffs(sam2 *Sample) int {
	//if there's any difference between the contained contigs, sum the rows as 'diffs'
	result := 0
	for leftKey := range sam1.contigs {
		_, ok := sam2.contigs[leftKey]
		if !ok {
			//sam2 contig doesn't contain sam1 contig
			//these count as a 'hit'
			result += len(sam1.contigs[leftKey])
			continue
		} else {
			//otherwise we add the diffs as normal
			sam1.contigs[leftKey].diff(sam2.contigs[leftKey])
		}
	}

	//same as before, except we don't re-calculate the diffs by missing contig
	for rightKey := range sam2.contigs {
		_, ok := sam1.contigs[rightKey]
		if !ok {
			//sam2 contig doesn't contain sam1 contig
			//these count as a 'hit'
			result += len(sam1.contigs[rightKey])
			continue
		}
	}

	return result


	
	//run diff against the matching

}

//removes the rows from the sample where the pos is in the list of positions
//positions must be sorted
func (sample *Sample)removeRowsOfPos(positions []uint32, wg *sync.WaitGroup) {
	defer wg.Done()
	for contigRef, rows := range sample.contigs {
		keepRows := []Row{}
		memIndex := 0
		for _, pos := range positions {
			for i:=memIndex; i<len(rows);{
				if rows[i].pos == pos {
					keepRows = append(keepRows, rows[i])
					memIndex++
					break
				}

				//if the given pos to rm is less than the row pos, move to next rmpos
				if pos < rows[i].pos {
					memIndex++
					break
				}

				// rm pos > rowpos
				memIndex++
				continue
			}
		}
		sample.contigs[contigRef] = keepRows
	}
}

//result is sorted
func collectSortedNonBasePositions(samples []Sample) (ret []uint32) {
	collect := make(map[uint32]bool)
	for _, sample := range samples {
		for _, rows := range sample.contigs {
			for _, row := range rows {
				if !validBase(row.base) {
					collect[row.pos] = true
				}
			}
		}
	}
	for k := range collect {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return
}

//takes two sets of rows, counts the "differences"
func (rows1 Rows) diff(rows2 Rows) int {
	//read from the left
	value := 0
	rightCounter := 0 
	for _, left := range rows1 {
		//while we haven't found a match or a larger contig/pos
		for j:=rightCounter; ; j++ {
			right := rows2[j]
			//if the first one on the right matches
			if right.pos > left.pos {
				//increment the val, not the counter
				//there is a row on the right that doesn't exist on the left
				value++
				break
				//continue to next row on the left
			}

			if right.pos < left.pos {
				//exists on left and not right
				//increment the right counter
				value++
				rightCounter = j
				rightCounter++
				continue
				//move along to the next right position
			}

			if right.pos == left.pos {
				//increment only if the base is different, move to next left one regardless
				if right.base != left.base {
					value++
					// fmt.Println("match except for base: ", sam1.name, left, string(left.base), sam2.name, right, string(right.base))
				}
				rightCounter++
				break
			}
			//we never get here
		}
		
	}
	sam2Len := len(rows2)
	
	//collect any hanging rows as 'counted' (don't exist on the left)
	if sam2Len > rightCounter {
		value += len(rows2[rightCounter:sam2Len])
	}
	return value
}


func outputResults(result [][]int, samples []Sample) {
	header := ","
	for _, sample := range samples {
		header += sample.name
		header += ","
	}
	//remove trailing comma
	header = strings.TrimSuffix(header, ",")
	header += "\n"
	fmt.Print(header)

	//each row: header then set of values
	for i, sample := range samples {
		strRow := sample.name + ","

		for _, resultElem := range result[i] {
			strRow += strconv.Itoa(resultElem)
			strRow += ","
		}
		
		strRow = strings.TrimSuffix(strRow, ",")
		strRow += "\n"
		fmt.Print(strRow)
	}
}