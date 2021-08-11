package main

import "testing"



func TestMain(t *testing.T) {
	main()
}

var diffTests = [][]string{
	[]string{

	},
}

func TestDiff(t *testing.T) {
	//should count a right tail (more on the right than the left)
	left := Sample{
		name: "A",
		rows: []Row{
			Row{
				contigRef: 0,
				pos: uint32(1),
				base: byte(1),
			},
		},
	}
	right := Sample{
		name: "B",
		rows: []Row{
			Row{
				contigRef: 0,
				pos: uint32(1),
				base: byte(1),
			},
			Row{
				contigRef: 0,
				pos: uint32(5),
				base: byte(1),
			},
		},
	}

	diff := left.diff(&right)
	// diff := right.diff(&left)
	if diff != 1 {
		t.Fatal("Must count the last missing on the right")
	}

	oppDiff := right.diff(&left)
	//should do this for a range of inputs
	if diff != oppDiff {
		t.Fatal("Should get the same results both ways")
	}

	matchBase := []Sample{
		Sample{
			name: "A",
			rows: []Row{
				Row{
					contigRef: 0,
					pos: 0,
					base: byte(0),
				},
			},
		},
		Sample{
			name: "B",
			rows: []Row{
				Row{
					contigRef: 0,
					pos: 0,
					base: byte(0),
				},
			},
		},
	}
	//if everything but bases match, should return 0
	matchBaseDiff := matchBase[0].diff(&matchBase[1])
	if matchBaseDiff == 0 {
		t.Fatal("bzzt")
	}


	matchBaseNo := []Sample{
		Sample{
			name: "A",
			rows: []Row{
				Row{
					contigRef: 0,
					pos: 0,
					base: byte(0),
				},
			},
		},
		Sample{
			name: "B",
			rows: []Row{
				Row{
					contigRef: 0,
					pos: 0,
					base: byte(1),
				},
			},
		},
	}
	//if everything but bases match, should return 0
	matchBaseDiffNo := matchBaseNo[0].diff(&matchBaseNo[1])
	if matchBaseDiffNo != 1 {
		t.Fatal("bzzot")
	}


	
}