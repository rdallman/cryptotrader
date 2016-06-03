package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kr/logfmt"
)

type line struct {
	Msg   string  `logfmt:msg`
	Price float64 `logfmt:price`

	Last float64 `logfmt:Last`
}

const fee = .0025

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "please provide a loggy file")
		os.Exit(1)
	}

	f, err := os.Open(args[0])
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

	for fast := 1; fast <= 50; fast++ {
		for slow := 1; slow <= 100; slow++ {
			for tick := 1; tick <= 120; tick++ {
				profit, f := tryEma(fast, slow, tick, lines)
				if profit > maxProfit {
					maxProfit, fees, bestF, bestS, bestT = profit, f, fast, slow, tick
				}
			}
		}
	}
	//fast, slow, tick := 1, 4, 50
	//profit := tryEma(fast, slow, tick, lines)
	//maxProfit, bestF, bestS, bestT = profit, fast, slow, tick
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
					*fees += .00002
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
