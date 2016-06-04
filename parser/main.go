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

	graph(avg[:], maxFast, maxSlow, maxTick, min)
}

func graph(matrix []float64, maxFast, maxSlow, maxTick int, min float64) {
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

	var maxProfit, fees float64
	var bestF, bestS, bestT int

	//for fast := 1; fast <= 50; fast++ {
	//for slow := 1; slow <= 100; slow++ {
	//for tick := 1; tick <= 120; tick++ {
	//profit, f := tryEma(fast, slow, tick, lines)
	//if profit > maxProfit {
	//maxProfit, fees, bestF, bestS, bestT = profit, f, fast, slow, tick
	//}
	//}
	//}
	//}
	fast, slow, tick := 10, 1, 5
	profit, fs := tryEma(fast, slow, tick, lines)
	maxProfit, fees, bestF, bestS, bestT = profit, fs, fast, slow, tick
	log.Printf("best: f=%d s=%d t=%d profit=%f fees=%f", bestF, bestS, bestT, maxProfit, fees)
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
		if *last == none { // set so we can see direction, TODO should base off open orders
			if f > s {
				*last = long
			} else {
				*last = short
			}
		} else {
			// TODO get out if below signal or below 0 / oppo for short... sit in cash yo
			if macd < 0 && macd < v && *last == long { // TODO fudge 25% ?
				*last = short

				if *lastBuy > 0 {
					p := price - *lastBuy
					//log.Printf("msg=SHORT msg2=LONGPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*fees += (*lastBuy * fee)
					*profit += p - (*lastBuy * fee)
				}
				*lastBuy = price
			} else if macd > 0 && macd > v && *last == short { // TODO fudge 25% ?
				*last = long

				if *lastBuy > 0 {
					p := *lastBuy - price
					//log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*fees += (*lastBuy * fee)
					*profit += p - (*lastBuy * fee)
				}
				*lastBuy = price
			}
		}
	}
}

type dir uint8

const (
	none dir = iota
	short
	long

	train      = 20 // TODO need to SMA until N instead of this
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
		if t < train {
			t++
			return notTrained
		}
		return avg
	}
}
