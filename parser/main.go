package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aybabtme/rgbterm"
	"github.com/kr/logfmt"
)

type line struct {
	Msg   string  `logfmt:msg`
	Price float64 `logfmt:price`

	Last float64 `logfmt:Last`
}

const fee = .0025

func main() {
	lb := flag.Bool("lb", false, "do leaderboard map")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "please provide a loggy file")
		os.Exit(1)
	}

	if *lb {
		leaderboard(args[0])
	} else {
		sim(args[0])
	}
}

func leaderboard(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(f)

	// NOTE: have to pull over these settings or there be dragons, we could search for them but blech
	const maxFast = 50
	const maxSlow = 100
	const maxTick = 24
	var min float64 = 1000000000000000000000000000000000000000000.
	var avg [maxFast * maxSlow * maxTick]float64
	var start bool
	for scanner.Scan() {
		b := scanner.Bytes()
		if !start {
			if bytes.HasPrefix(b, []byte("leaderboard")) {
				start = true
				scanner.Scan() // skip new line
			}
			continue
		}
		if bytes.Contains(b, []byte("best")) {
			break // over
		}

		m := bytes.Split(b, []byte(":"))
		rank, err := strconv.ParseFloat(string(bytes.TrimSpace(m[1])), 64)
		errNil(err)
		if rank < min {
			min = rank
		}

		list := bytes.Split(m[0], []byte("/"))
		fast, err := strconv.Atoi(string(bytes.TrimSpace(list[0])))
		errNil(err)
		slow, err := strconv.Atoi(string(bytes.TrimSpace(list[1])))
		errNil(err)
		tick, err := strconv.Atoi(string(bytes.TrimSpace(list[2])))
		errNil(err)

		avg[(tick-1)+(slow-1)*maxTick+(fast-1)*maxSlow*maxTick] = rank
	}

	graph2d(avg[:], maxFast, maxSlow, maxTick, min)
}

func graph2d(matrix []float64, maxFast, maxSlow, maxTick int, min float64) {
	// Print top axis
	fmt.Printf("f/s") // skip axis width
	for x := 1; x <= maxTick; x++ {
		fmt.Printf(" %-6d", x)
	}
	fmt.Println()

	// Print matrix
	for fast := 1; fast <= maxFast; fast++ {
		for slow := 1; slow <= maxSlow; slow++ {
			// Print left axis
			fmt.Printf("%3d/%-3d ", fast, slow)
			for tick := 1; tick <= maxTick; tick++ { // up to 2 hours
				p := matrix[(tick-1)+(slow-1)*maxTick+(fast-1)*maxSlow*maxTick]
				r, g, b := color(p, min, min*2)                                              // bluer is better
				fmt.Printf("%s ", rgbterm.String(fmt.Sprintf("%-6.f", p), r, g, b, 0, 0, 0)) // Multiply by some constant to make it human readable
			}
			fmt.Println()
		}
	}
}

func color(v, vmin, vmax float64) (rd, gn, bl uint8) {
	r, g, b := 1.0, 1.0, 1.0 // white
	if v < vmin {
		v = vmin
	}
	if v > vmax {
		v = vmax
	}
	dv := vmax - vmin
	if v < (vmin + 0.25*dv) {
		r = 0
		g = 4 * (v - vmin) / dv
	} else if v < (vmin + 0.5*dv) {
		r = 0
		b = 1 + 4*(vmin+0.25*dv-v)/dv
	} else if v < (vmin + 0.75*dv) {
		r = 4 * (v - vmin - 0.5*dv) / dv
		b = 0
	} else {
		g = 1 + 4*(vmin+0.75*dv-v)/dv
		b = 0
	}
	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func errNil(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func sim(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(f)

	var lines []line
	for scanner.Scan() {
		var l line
		if err := logfmt.Unmarshal(scanner.Bytes(), &l); err != nil {
			fmt.Fprintln(os.Stderr, err) // sweet baby jesus why '14t'
			continue
		}
		lines = append(lines, l)
	}

	var maxProfit, fees, first float64
	var bestF, bestS, bestT int

	//const maxFast = 50
	//const maxSlow = 100
	//const maxTick = 36
	////maxSig := 10
	////sig := 2

	//// y = tick, x = fast/slow ema combos
	//matrix := make([][]float64, maxFast*maxSlow)

	//var cur, avg [maxFast * maxSlow * maxTick]rank

	//var left [][3]int
	//for fast := 1; fast <= maxFast; fast++ {
	//for slow := 1; slow <= maxSlow; slow++ {
	//left = append(left, [3]int{fast, slow, 0})
	//for tick := 1; tick <= maxTick; tick++ {
	//p := [3]int{fast, slow, tick}
	//avg[(tick-1)+(slow-1)*maxTick+(fast-1)*maxSlow*maxTick] = rank{p: p}
	//}
	//}
	//}

	//trials := 180
	//shortest := 180
	//_ = shortest

	//first = lines[0].Last

	//for days := shortest; days <= trials; days++ {
	//log.Println("trial day:", days)

	//var i int
	//for fast := 1; fast <= maxFast; fast++ {
	//for slow := 1; slow <= maxSlow; slow++ {
	//matrix[i] = make([]float64, maxTick)
	//for tick := 1; tick <= maxTick; tick++ { // up to 2 hours
	//profit, f := tryEma(fast, slow, tick, lines)
	//p := 100 * (profit / 1)
	//matrix[i][tick-1] = p
	//if profit > maxProfit {
	//maxProfit, fees, bestF, bestS, bestT = profit, f, fast, slow, tick
	//}
	//cur[(tick-1)+(slow-1)*maxTick+(fast-1)*maxSlow*maxTick] = rank{p: [3]int{fast, slow, tick}, a: p}
	//}
	//i++
	//}
	//}

	//// sort by profits, highest profit first (= lowest rank = best)
	//sort.Stable(profits(cur[:]))
	//for i, c := range cur {
	//// track total ranks thus far, average later
	//avg[(c.p[2]-1)+(c.p[1]-1)*maxTick+(c.p[0]-1)*maxSlow*maxTick].a += float64(i + 1)
	//}
	//}

	//for i := range avg {
	//avg[i].a /= float64(trials - shortest + 1)
	//}

	//sort.Stable(ranks(avg[:]))

	//fmt.Println("leaderboard: [fast/slow/tick]: [avg rank]")
	//fmt.Println()
	//for i, a := range avg {
	//fmt.Printf("%10d: %3d/%3d/%3d: %9.3f\n", i+1, a.p[0], a.p[1], a.p[2], a.a)
	//}

	fast, slow, tick := 47, 81, 135
	profit, fe := tryEma(fast, slow, tick, lines)
	maxProfit, fees, bestF, bestS, bestT = profit, fe, fast, slow, tick
	log.Printf("best: f=%d s=%d t=%d profit%%=%f profit=%f fees=%f price=%f", bestF, bestS, bestT, 100*(maxProfit/1), maxProfit, fees, first)

	// find best box, by fast ema
	//var maxBox float64
	//var top, leftN, bottom, right int
	//for i := 0; i < len(matrix)/(maxSlow); i++ {
	//maxSubBox, t, l, b, r := max_contiguous_submatrix(matrix[i*maxSlow : (i+1)*maxSlow])
	//t += (i * maxSlow)
	//b += (i * maxSlow)
	//log.Printf("box: maxBox=%f top=%d/%d left=%d bottom=%d/%d right=%d", maxSubBox, left[t][0], left[t][1], l, left[b][0], left[b][1], r)
	//if maxSubBox > maxBox {
	//maxBox, top, leftN, bottom, right = maxSubBox, t, l, b, r
	//}
	//}

	//log.Printf("best box: maxBox=%f top=%d/%d left=%d bottom=%d/%d right=%d", maxBox, left[top][0], left[top][1], leftN, left[bottom][0], left[bottom][1], right)
	//graph(matrix, left, 100*(maxProfit/1))
}

type rank struct {
	p [3]int
	a float64
}

type ranks []rank

func (r ranks) Len() int           { return len(r) }
func (r ranks) Less(i, j int) bool { return r[i].a < r[j].a } // lowest first
func (r ranks) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

type profits []rank

func (r profits) Len() int           { return len(r) }
func (r profits) Less(i, j int) bool { return r[i].a > r[j].a } // highest first
func (r profits) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func graph(matrix [][]float64, left [][3]int, max float64) {
	// Print top axis
	fmt.Printf("f/s") // skip axis width
	for x := 1; x <= len(matrix[0]); x++ {
		fmt.Printf(" %-6d ", x)
	}
	fmt.Println()

	// Print matrix
	for a, col := range matrix {
		// Print left axis
		fmt.Printf("%3d/%-3d ", left[a][0], left[a][1])
		// Print row of values
		for _, count := range col {
			p := count
			r, g, b := color(p, max*.25, max*.5)
			fmt.Printf("%s ", rgbterm.String(fmt.Sprintf("%-6.f", p), r, g, b, 0, 0, 0)) // Multiply by some constant to make it human readable
		}
		fmt.Println()
	}
}

// TODO this is max sum, we want largest, highest avg @ some weight ?

// O(N^3) find largest sum submatrix
//
// NOTE: maxRows means this will find the largest NxM matrix where
// N will be no greater than maxRows. toggle off to remove
func max_contiguous_submatrix(m [][]float64) (m64 float64, t, l, b, r int) {
	maxRows := 10
	//maxCols := 12 // TODO have to be smahter in kandane

	rows := len(m)
	cols := len(m[0])

	vps := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		vps[i] = make([]float64, cols)
	}

	for j := 0; j < cols; j++ {
		vps[0][j] = m[0][j]
		for i := 1; i < rows; i++ {
			vps[i][j] = vps[i-1][j] + m[i][j]
		}
	}

	max, top, left, bottom, right := m[0][0], 0, 0, 0, 0
	// max = [m[0][0],0,0,0,0] // this is the result, stores [max,top,left,bottom,right]

	// these arrays are used over Kandane
	sum := make([]float64, cols) // obvious sum array used in Kandane
	pos := make([]int, cols)     // keeps track of the beginning position for the max subseq ending in j

	for i := 0; i < rows; i++ {
		for k := i; k < rows; k++ {
			if k-i > maxRows { // NOTE: max N to find a tighter box
				break // move down the rows
			}

			// Kandane over all columns with the i..k rows
			// clean both the sum and pos arrays for the upcoming kandane
			for z := 0; z < cols; z++ {
				sum[z] = 0
				pos[z] = 0
			}
			local_max := 0 //  we keep track of the position of the max value over each Kandane's execution
			// notice that we do not keep track of the max value, but only its position
			var d float64
			if i > 0 {
				d = vps[i-1][0]
			}
			sum[0] = vps[k][0] - d
			for j := 1; j < cols; j++ {
				var d float64
				if i > 0 {
					d = vps[i-1][j]
				}
				value := vps[k][j] - d
				if sum[j-1] > 0 {
					sum[j] = sum[j-1] + value
					pos[j] = pos[j-1]
				} else {
					sum[j] = value
					pos[j] = j
				}
				if sum[j] > sum[local_max] {
					local_max = j
				}
			}
			// Kandane ends here

			// Here's the key thing
			// If the max value obtained over the past kandane's execution is larger than
			// the current maximum, then update the max array with sum and bounds
			if sum[local_max] > max {
				// sum[local_max] is the new max value
				// the corresponding submatrix goes from rows i..k.
				// and from columns pos[local_max]..local_max
				// the array below contains [max_sum,top,left,bottom,right]
				max, top, left, bottom, right = sum[local_max], i, pos[local_max], k, local_max
			}
		}
	}

	return max, top, left, bottom, right
}

func tryEma(fast, slow, tickC int, lines []line) (float64, float64) {

	// TODO get open margin orders, set dir based on that..

	emaFast := ema(fast)
	emaSlow := ema(slow)
	signal := ema(2)
	_ = signal
	var lastDir dir
	var tick int
	var lastBuy, profit, fees float64

	for _, l := range lines {
		if l.Last == 0 {
			continue
		}
		tick++

		// profit calc from log of prices
		if tick%tickC == 0 {
			tradeMACD(l.Last, &lastBuy, &profit, &fees, &lastDir, emaFast, emaSlow, signal)
			//tradeEMA(l.Last, &lastBuy, &profit, &fees, &lastDir, emaFast, emaSlow)
		}
	}

	return profit, fees
}

func tradeEMA(price float64, lastBuy, profit, fees *float64, last *dir, emaFast, emaSlow func(float64) float64) {
	f := emaFast(price)
	s := emaSlow(price)
	if f != notTrained && s != notTrained {
		if *last == none { // set so we can see direction
			if f > s {
				*last = long
			} else {
				*last = short
			}
		} else {
			if f < s && *last == long {
				*last = short

				if *lastBuy > 0 {
					p := price - *lastBuy
					//				log.Printf("msg=SHORT msg2=LONGPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*fees += .00002
					*fees += (*lastBuy * fee)
					*profit += p - (*lastBuy * fee)
				}
				*lastBuy = price
			} else if f > s && *last == short {
				*last = long

				if *lastBuy > 0 {
					p := *lastBuy - price
					//					log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*fees += (*lastBuy * fee)
					*profit += p - (*lastBuy * fee)
				}
				*lastBuy = price
			}
		}
	}
}

func tradeMACD(price float64, lastBuy, profit, fees *float64, last *dir, emaFast, emaSlow, signal func(float64) float64) {
	f := emaFast(price)
	s := emaSlow(price)

	// MACD Line: (12-day EMA - 26-day EMA)
	// Signal Line: 9-day EMA of MACD Line
	macd := f - s
	v := signal(macd)

	if f != notTrained && s != notTrained {
		if *last == long && ( /*macd < 0 ||*/ macd < v) {
			// close long
			mult := 1. // + *profit // NOTE: add profit back for compound
			p := mult * ((price - *lastBuy) / *lastBuy)
			// NOTE compound ends up weighting later profits higher, which sucks (but shiny)

			//log.Printf("msg=LONGPROFITS buy=%f price=%f profit=%f gross_profit=%f net_profit=%f fee=%f", *lastBuy, price, p, *profit+p, *profit+p-f, f)
			f := fee * mult
			*fees += f
			*profit += p - f
			*last = none
		} else if *last == short && ( /*macd > 0 ||*/ macd > v) {
			// close short
			mult := 1. // + *profit // NOTE: add profit back for compound

			p := mult * ((*lastBuy - price) / *lastBuy)
			// change profit in terms of eth_btc to be in terms of eth

			// log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f gross_profit=%f net_profit=%f fee=%f", *lastBuy, price, p, *profit+p, *profit+p-f, f)
			const lending = .0002
			f := (fee * mult) + (lending * mult)

			*fees += f
			*profit += p - f
			*last = none
		}

		// open new ones, if necessary
		if *last == none && /*macd < 0 &&*/ macd < v /*&&((v - macd) / v) > .01*/ { // TODO fudge 25% ? TODO confirm < 0 ?
			*last = short
			*lastBuy = price
		} else if *last == none && /*macd > 0 &&*/ macd > v /*&&((macd - v) / macd) > .01*/ { // TODO fudge 25% ? TODO confirm > 0 ?
			*last = long
			*lastBuy = price
		}
	}
}

type dir uint8

const (
	none dir = iota
	short
	long

	notTrained = -1
)

// EMA = Price(t) * k + EMA(y) * (1 – k)
// k = 2/(N+1)
func ema(n int) func(float64) float64 {
	var avg float64
	k := 2 / (float64(n) + 1)
	var t int
	return func(f float64) float64 {
		avg = f*k + avg*(1-k)
		if t < n {
			t++
			return notTrained
		}
		return avg
	}
}
