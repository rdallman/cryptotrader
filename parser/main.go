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

	for fast := 0; fast < 20; fast++ {
		for slow := 0; slow < 50; slow++ {
			for tick := 0; tick < 60; tick++ {
				f.Seek(0, 0)
				tryEma(fast, slow, tick, f)
			}
		}
	}
}

func tryEma(fast, slow, tickC int, f *os.File) {
	scanner := bufio.NewScanner(f)

	lines := make(chan line, 1024)
	go func() {
		for scanner.Scan() {
			var l line
			if err := logfmt.Unmarshal(scanner.Bytes(), &l); err != nil {
				fmt.Fprintln(os.Stderr, err) // sweet baby jesus why '14t'
				continue
			}
			lines <- l
		}
		close(lines)
	}()

	// TODO get open margin orders, set dir based on that..

	emaFast := ema(13)
	emaSlow := ema(41)
	var lastDir dir
	var tick int
	var lastBuy, profit float64

	for l := range lines {
		//if l.Msg == "" { continue }
		//if lastPrice == 0 { // first trade
		//lastPrice = l.Price
		//continue
		//}

		//last := lastPrice
		//// profit calc w/i log
		//var profit float64
		//switch l.Msg {
		//case "SHORT": // were long, now short
		//profit = l.Price - lastPrice
		//lastPrice = l.Price
		//case "LONG": // were short, now long
		//profit = lastPrice - l.Price
		//lastPrice = l.Price
		//}
		//profits += profit

		//log.Println("money", "last", last, "price", l.Price, "pos", l.Msg, "profit", profit, "total_profit", profits) // TODO time held

		if l.Last == 0 {
			continue
		}
		tick++

		// profit calc from log of prices
		if tick%tickC == 0 {
			trade(l.Last, &lastBuy, &profit, &lastDir, emaFast, emaSlow)
		}
	}
}

func trade(price float64, lastBuy, profit *float64, last *dir, emaFast, emaSlow func(float64) float64) {
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
					log.Printf("msg=SHORT msg2=LONGPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*profit += p
				}
				*lastBuy = price
			} else if f > s && *last == short {
				*last = long

				if *lastBuy > 0 {
					p := *lastBuy - price
					log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*profit += p
				}
				*lastBuy = price
			}

			//switch l.Msg {
			//case "SHORT": // were long, now short
			//profit = l.Price - lastPrice
			//lastPrice = l.Price
			//case "LONG": // were short, now long
			//profit = lastPrice - l.Price
			//lastPrice = l.Price
			//}

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
