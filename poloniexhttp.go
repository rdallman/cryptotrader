package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/aybabtme/rgbterm"
)

const (
	POLONIEX_API_URL                = "https://poloniex.com"
	POLONIEX_API_TRADING_ENDPOINT   = "tradingApi"
	POLONIEX_API_VERSION            = "1"
	POLONIEX_BALANCES               = "returnBalances"
	POLONIEX_BALANCES_COMPLETE      = "returnCompleteBalances"
	POLONIEX_DEPOSIT_ADDRESSES      = "returnDepositAddresses"
	POLONIEX_GENERATE_NEW_ADDRESS   = "generateNewAddress"
	POLONIEX_DEPOSITS_WITHDRAWALS   = "returnDepositsWithdrawals"
	POLONIEX_ORDERS                 = "returnOpenOrders"
	POLONIEX_TRADE_HISTORY          = "returnTradeHistory"
	POLONIEX_ORDER_BUY              = "buy"
	POLONIEX_ORDER_SELL             = "sell"
	POLONIEX_ORDER_CANCEL           = "cancelOrder"
	POLONIEX_ORDER_MOVE             = "moveOrder"
	POLONIEX_WITHDRAW               = "withdraw"
	POLONIEX_FEE_INFO               = "returnFeeInfo"
	POLONIEX_AVAILABLE_BALANCES     = "returnAvailableAccountBalances"
	POLONIEX_TRADABLE_BALANCES      = "returnTradableBalances"
	POLONIEX_TRANSFER_BALANCE       = "transferBalance"
	POLONIEX_MARGIN_ACCOUNT_SUMMARY = "returnMarginAccountSummary"
	POLONIEX_MARGIN_BUY             = "marginBuy"
	POLONIEX_MARGIN_SELL            = "marginSell"
	POLONIEX_MARGIN_POSITION        = "getMarginPosition"
	POLONIEX_MARGIN_POSITION_CLOSE  = "closeMarginPosition"
	POLONIEX_CREATE_LOAN_OFFER      = "createLoanOffer"
	POLONIEX_CANCEL_LOAN_OFFER      = "cancelLoanOffer"
	POLONIEX_OPEN_LOAN_OFFERS       = "returnOpenLoanOffers"
	POLONIEX_ACTIVE_LOANS           = "returnActiveLoans"
	POLONIEX_AUTO_RENEW             = "toggleAutoRenew"

	fee = .0025
)

type Poloniex struct {
	Name                    string
	Enabled                 bool
	Verbose                 bool
	Websocket               bool
	RESTPollingDelay        time.Duration
	AuthenticatedAPISupport bool
	AccessKey, SecretKey    string
	Fee                     float64
	BaseCurrencies          []string
	AvailablePairs          []string
	EnabledPairs            []string
}

type PoloniexTicker struct {
	Last          float64 `json:"last,string"`
	LowestAsk     float64 `json:"lowestAsk,string"`
	HighestBid    float64 `json:"highestBid,string"`
	PercentChange float64 `json:"percentChange,string"`
	BaseVolume    float64 `json:"baseVolume,string"`
	QuoteVolume   float64 `json:"quoteVolume,string"`
	IsFrozen      int     `json:"isFrozen,string"`
	High24Hr      float64 `json:"high24hr,string"`
	Low24Hr       float64 `json:"low24hr,string"`
}

func (p *Poloniex) SetDefaults() {
	p.Name = "Poloniex"
	p.Enabled = false
	p.Fee = 0
	p.Verbose = false
	p.Websocket = false
	p.RESTPollingDelay = 10
}

func (p *Poloniex) GetName() string {
	return p.Name
}

func (p *Poloniex) SetEnabled(enabled bool) {
	p.Enabled = enabled
}

func (p *Poloniex) IsEnabled() bool {
	return p.Enabled
}

func (p *Poloniex) Setup(exch Exchanges) {
	if !exch.Enabled {
		p.SetEnabled(false)
	} else {
		p.Enabled = true
		p.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		p.SetAPIKeys(exch.APIKey, exch.APISecret)
		p.RESTPollingDelay = exch.RESTPollingDelay
		p.Verbose = exch.Verbose
		p.Websocket = exch.Websocket
		p.BaseCurrencies = SplitStrings(exch.BaseCurrencies, ",")
		p.AvailablePairs = SplitStrings(exch.AvailablePairs, ",")
		p.EnabledPairs = SplitStrings(exch.EnabledPairs, ",")
	}
}

func (p *Poloniex) Start() {
	go p.Run()
}

func (p *Poloniex) SetAPIKeys(apiKey, apiSecret string) {
	p.AccessKey = apiKey
	p.SecretKey = apiSecret
}

func (p *Poloniex) GetFee() float64 {
	return p.Fee
}

func (p *Poloniex) Run() {
	if p.Verbose {
		log.Printf("%s Websocket: %s (url: %s).\n", p.GetName(), IsEnabled(p.Websocket), POLONIEX_WEBSOCKET_ADDRESS)
		log.Printf("%s polling delay: %ds.\n", p.GetName(), p.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", p.GetName(), len(p.EnabledPairs), p.EnabledPairs)
	}

	if p.Websocket {
		go p.WebsocketClient()
	}

	currency := "BTC_ETH"
	//	for p.Enabled {

	//	for _, currency := range p.EnabledPairs {
	//go func() {
	//log.Printf("Poloniex=%s Last=%f High=%f Low=%f Volume=%f Ask=%f Bid=%f EMA13=%f EMA41=%f\n", currency, t.Last, t.High24Hr, t.Low24Hr, t.QuoteVolume, t.LowestAsk, t.HighestBid, f, s)
	//}()

	// TODO sims lay below, extract somehow
	//p.tryAll(currency)

	//for _, days := range []int{7, 14, 21, 30, 60, 90, 120, 150} {
	//p.tryOne(currency, days, 5, 1, 95)
	//p.tryOne(currency, days, 42, 48, 90)
	//p.tryOne(currency, days, 9, 90, 90)
	//p.tryOne(currency, days, 29, 37, 90)
	//p.tryOne(currency, days, 17, 30, 120)
	//p.tryOne(currency, days, 22, 31, 120)
	//p.tryOne(currency, days, 47, 56, 60)
	//p.tryOne(currency, days, 47, 48, 60)
	//p.tryOne(currency, days, 47, 81, 135)
	//p.tryOne(currency, days, 47, 81, 120)
	//p.tryOne(currency, days, 50, 85, 120)
	//p.tryOne(currency, days, 36, 44, 130)
	//p.tryOne(currency, days, 50, 92, 150)
	//p.tryOne(currency, days, 50, 80, 150)
	//p.tryOne(currency, days, 39, 50, 125)
	//}

	//time.Sleep(time.Second * p.RESTPollingDelay)
	//}
	//	}
}

func (p *Poloniex) realTrade(currency string) {
	acc, err := p.GetMarginAccountSummary()
	if err != nil {
		log.Fatalf("couldn't get margin account info: %v", err)
	}

	log.Println("account balance: ", acc.TotalValue)
	log.Println("account info: ", acc)

	openI, err := p.GetMarginPosition(currency)
	if err != nil {
		log.Fatalf("couldn't get margin account info: %v", err)
	}
	open := openI.(PoloniexMarginPosition)

	var pos dir
	switch open.Type {
	case "none":
		pos = none
	case "long":
		pos = long
	case "short":
		pos = short
	}
	lastBuy := open.BasePrice

	log.Printf("open pos: type=%s price=%f amount=%f total=%f p/l=%f lending_fees=%f", open.Type, open.BasePrice, open.Amount, open.Total, open.ProfitLoss, open.LendingFees)

	var profit, fees float64
	fast, slow, sig := 17, 30, 2 // TODO make these configuramable ?
	const candle = 7200          // 2hr candle; candlestick period in seconds; valid values are 300, 900, 1800, 7200, 14400, and 86400
	emaFast := ema(fast)
	emaSlow := ema(slow)
	signal := ema(sig)

	// get enough back data to seed the fast / slow emas so we can start trading..
	lastData := time.Now()
	timemachine := time.Duration(int(math.Max(float64(fast), float64(slow)))) * candle * time.Second
	start := strconv.Itoa(int(lastData.Add(-timemachine).Unix()))
	end := strconv.Itoa(int(lastData.Unix()))
	period := strconv.Itoa(candle)
	chart, err := p.GetChartData(currency, start, end, period)
	if err != nil {
		log.Fatalf("issue going back in time; check the flux capacitor. err:", err)
	}

	// initialize all the things
	for _, pt := range chart {
		tradeMACD(pt.Close, &lastBuy, &profit, &fees, &pos, emaFast, emaSlow, signal)
	}

	var price float64
	for ; ; time.Sleep(candle * time.Second) {
		var errCount int
		// retry for 2 minutes, and then bail
		try := lastData.Add(candle * time.Second)
		tryS := strconv.Itoa(int(try.Unix()))
		for ; ; time.Sleep(4 * time.Second) {
			// just get one data point
			// TODO does this work? ticker will be imprecise, esp. with retries. could get 2 pts and discard 1
			c, err := p.GetChartData(currency, tryS, tryS, period)
			if err != nil {
				errCount++
				if errCount > 30 { // TODO assert that sleep * this is < fast ema
					log.Fatalf("couldn't get price, maybe polo is down? err:", err)
				}
				continue
			}
			if len(c) < 1 {
				log.Fatalf("fix the fuckin robot man")
			}
			price = c[0].Close
			lastData = try
			break
		}

		tradeMACD(price, &lastBuy, &profit, &fees, &pos, emaFast, emaSlow, signal)
		// TODO execute the trade
	}
}

func (p *Poloniex) tryOne(currency string, days, fast, slow, tick int) {
	sig := 2
	tick /= 5

	start := strconv.Itoa(int(time.Now().Add(-24 * time.Hour * time.Duration(days)).Unix()))
	end := strconv.Itoa(int(time.Now().Unix()))
	period := "300" // min allowed is 5 min candles; use for everything
	c, err := p.GetChartData(currency, start, end, period)
	if err != nil {
		log.Fatal("fucked up chart data:", err)
	}

	var first float64
	if len(c) > 0 {
		first = c[0].Close
	} else {
		log.Print("no chart data, this will prove futile")
	}

	profit, f := tryEma(fast, slow, sig, tick, c)

	log.Printf("%s profit: f=%d s=%d t=%d profit%%=%f profit=%f fees=%f price=%f", currency, fast, slow, tick*5, 100*(profit/1), profit, f, first)

	// TODO print out each trade
}

func (p *Poloniex) tryAll(currency string) {
	var maxProfit, fees, first float64
	var bestF, bestS, bestT int

	const maxFast = 50
	const maxSlow = 100
	const maxTick = 36
	//maxSig := 10
	sig := 2

	// y = tick, x = fast/slow ema combos
	matrix := make([][]float64, maxFast*maxSlow)

	var cur, avg [maxFast * maxSlow * maxTick]rank

	var left [][3]int
	for fast := 1; fast <= maxFast; fast++ {
		for slow := 1; slow <= maxSlow; slow++ {
			left = append(left, [3]int{fast, slow, 0})
			for tick := 1; tick <= maxTick; tick++ {
				p := [3]int{fast, slow, tick}
				avg[(tick-1)+(slow-1)*maxTick+(fast-1)*maxSlow*maxTick] = rank{p: p}
			}
		}
	}

	trials := 35
	shortest := 35
	_ = shortest

	for days := shortest; days <= trials; days++ {
		log.Println("trial day:", days)
		start := strconv.Itoa(int(time.Now().Add(-24 * time.Hour * time.Duration(days)).Unix()))
		end := strconv.Itoa(int(time.Now().Unix()))
		period := "300" // min allowed is 5 min candles; use for everything
		c, err := p.GetChartData(currency, start, end, period)
		if err != nil {
			log.Fatal("fucked up chart data:", err)
		}

		if len(c) > 0 {
			first = c[0].Close
		} else {
			log.Print("no chart data, this will prove futile")
		}

		var i int
		for fast := 1; fast <= maxFast; fast++ {
			for slow := 1; slow <= maxSlow; slow++ {
				matrix[i] = make([]float64, maxTick)
				for tick := 1; tick <= maxTick; tick++ { // up to 2 hours
					profit, f := tryEma(first, fast, slow, sig, tick, c)
					p := 100 * (profit / 1)
					matrix[i][tick-1] = p
					if profit > maxProfit {
						maxProfit, fees, bestF, bestS, bestT = profit, f, fast, slow, tick
					}
					cur[(tick-1)+(slow-1)*maxTick+(fast-1)*maxSlow*maxTick] = rank{p: [3]int{fast, slow, tick}, a: p}
				}
				i++
			}
		}

		// sort by profits, highest profit first (= lowest rank = best)
		sort.Stable(profits(cur[:]))
		for i, c := range cur {
			// track total ranks thus far, average later
			avg[(c.p[2]-1)+(c.p[1]-1)*maxTick+(c.p[0]-1)*maxSlow*maxTick].a += float64(i + 1)
		}
	}

	for i := range avg {
		avg[i].a /= float64(trials - shortest + 1)
	}

	sort.Stable(ranks(avg[:]))

	fmt.Println("leaderboard: [fast/slow/tick]: [avg rank]")
	fmt.Println()
	for i, a := range avg {
		fmt.Printf("%10d: %3d/%3d/%3d: %9.3f\n", i+1, a.p[0], a.p[1], a.p[2], a.a)
	}

	//fast, slow, tick := 13, 41, 12
	//profit, f := tryEma(fast, slow, tick, c)
	//maxProfit, fees, bestF, bestS, bestT = profit, f, fast, slow, tick
	log.Printf("%s best: f=%d s=%d t=%d profit%%=%f profit=%f fees=%f price=%f", currency, bestF, bestS, bestT*5, 100*(maxProfit/1), maxProfit, fees, first)

	// find best box, by fast ema
	var maxBox float64
	var top, leftN, bottom, right int
	for i := 0; i < len(matrix)/(maxSlow); i++ {
		maxSubBox, t, l, b, r := max_contiguous_submatrix(matrix[i*maxSlow : (i+1)*maxSlow])
		t += (i * maxSlow)
		b += (i * maxSlow)
		log.Printf("%s box: maxBox=%f top=%d/%d left=%d bottom=%d/%d right=%d", currency, maxSubBox, left[t][0], left[t][1], l, left[b][0], left[b][1], r)
		if maxSubBox > maxBox {
			maxBox, top, leftN, bottom, right = maxSubBox, t, l, b, r
		}
	}

	log.Printf("%s best box: maxBox=%f top=%d/%d left=%d bottom=%d/%d right=%d", currency, maxBox, left[top][0], left[top][1], leftN, left[bottom][0], left[bottom][1], right)
	graph(matrix, left, 100*(maxProfit/1))
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

func tryEma(fast, slow, sig, tickC int, lines []PoloniexChartData) (float64, float64) {
	// TODO get open margin orders, set dir based on that..

	emaFast := ema(fast)
	emaSlow := ema(slow)
	signal := ema(sig)
	_ = signal
	var lastDir dir
	var tick int
	var lastBuy, profit, fees float64

	for _, l := range lines {
		tick++

		// profit calc from log of prices
		if tick%tickC == 0 {
			tradeMACD(l.Close, &lastBuy, &profit, &fees, &lastDir, emaFast, emaSlow, signal)
			//tradeEMA(l.Close, &lastBuy, &profit, &fees, &lastDir, emaFast, emaSlow)
		}
	}
	return profit, fees
}

func tradeMACD(price float64, lastBuy, profit, fees *float64, last *dir, emaFast, emaSlow, signal func(float64) float64) {
	f := emaFast(price)
	s := emaSlow(price)

	// MACD Line: (12-day EMA - 26-day EMA)
	// Signal Line: 9-day EMA of MACD Line
	macd := f - s
	v := signal(macd)

	if f != notTrained && s != notTrained {
		// TODO set direction should be based off open orders
		// go long if macd > 0 && macd > v
		// close long if macd < 0 || macd < v
		// go short if macd < 0 && macd < v
		// close short if macd > 0 || macd > v

		// TODO calculate exposure

		// close order first
		if *lastBuy > 0 && *last == long && ( /*macd < 0 ||*/ macd < v) {
			// close long
			mult := 1. + *profit // NOTE: add profit back for compound
			p := mult * ((price - *lastBuy) / *lastBuy)
			// NOTE compound ends up weighting later profits higher, which sucks (but shiny)

			//log.Printf("msg=LONGPROFITS buy=%f price=%f profit=%f gross_profit=%f net_profit=%f fee=%f", *lastBuy, price, p, *profit+p, *profit+p-f, f)
			f := fee * mult
			*fees += f
			*profit += p - f
			*last = none
		} else if *lastBuy > 0 && *last == short && ( /*macd > 0 ||*/ macd > v) {
			// close short
			mult := 1. // + *profit // NOTE: add profit back for compound

			p := mult * ((*lastBuy - price) / *lastBuy)
			// change profit in terms of eth_btc to be in terms of eth

			// log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f gross_profit=%f net_profit=%f fee=%f", *lastBuy, price, p, *profit+p, *profit+p-f, f)
			//const lending = .0002
			f := (fee * mult) + (.0002 * mult)

			*fees += f
			*profit += p - f
			*last = none
		}

		if *last == none && *lastBuy == 0 { // nope
			if macd < v {
				*last = short
			} else {
				*last = long
			}
			// don't set price, just wait for a crossover
		}

		// open new ones, if necessary
		if /*macd < 0 &&*/ macd < v /*&&((v - macd) / v) > .01*/ { // TODO fudge 25% ? TODO confirm < 0 ?
			// only go short if on the first trade, we were looking for a short xover or we were in cash.
			// i.e. don't make the first trade until the first xover...
			if *last == none || (*lastBuy == 0 && *last == long) {
				*lastBuy = price
				*last = short
			}
		} else if /*macd > 0 &&*/ macd > v /*&&((macd - v) / macd) > .01*/ { // TODO fudge 25% ? TODO confirm > 0 ?
			if *last == none || (*lastBuy == 0 && *last == short) {
				*last = long
				*lastBuy = price
			}
		}
	}
}

// TODO needs updating... was worse than macd
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

					// log.Printf("msg=SHORT msg2=LONGPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
					*fees += (*lastBuy * fee)
					*profit += p - (*lastBuy * fee)
				}
				*lastBuy = price
			} else if f > s && *last == short {
				*last = long

				if *lastBuy > 0 {
					p := *lastBuy - price
					// log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
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
	cash

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

func (p *Poloniex) GetTicker() (map[string]PoloniexTicker, error) {
	type response struct {
		Data map[string]PoloniexTicker
	}

	resp := response{}
	path := fmt.Sprintf("%s/public?command=returnTicker", POLONIEX_API_URL)
	err := SendHTTPGetRequest(path, true, &resp.Data)

	if err != nil {
		return resp.Data, err
	}
	return resp.Data, nil
}

func (p *Poloniex) GetVolume() (interface{}, error) {
	var resp interface{}
	path := fmt.Sprintf("%s/public?command=return24hVolume", POLONIEX_API_URL)
	err := SendHTTPGetRequest(path, true, &resp)

	if err != nil {
		return resp, err
	}
	return resp, nil
}

type PoloniexOrderbook struct {
	Asks     [][]interface{} `json:"asks"`
	Bids     [][]interface{} `json:"bids"`
	IsFrozen string          `json:"isFrozen"`
}

//TO-DO: add support for individual pair depth fetching
func (p *Poloniex) GetOrderbook(currencyPair string, depth int) (map[string]PoloniexOrderbook, error) {
	type Response struct {
		Data map[string]PoloniexOrderbook
	}

	vals := url.Values{}
	vals.Set("currencyPair", currencyPair)

	if depth != 0 {
		vals.Set("depth", strconv.Itoa(depth))
	}

	resp := Response{}
	path := fmt.Sprintf("%s/public?command=returnOrderBook&%s", POLONIEX_API_URL, vals.Encode())
	err := SendHTTPGetRequest(path, true, &resp.Data)

	if err != nil {
		return resp.Data, err
	}
	return resp.Data, nil
}

type PoloniexTradeHistory struct {
	GlobalTradeID int64   `json:"globalTradeID"`
	TradeID       int64   `json:"tradeID"`
	Date          string  `json:"date"`
	Type          string  `json:"type"`
	Rate          float64 `json:"rate,string"`
	Amount        float64 `json:"amount,string"`
	Total         float64 `json:"total,string"`
}

func (p *Poloniex) GetTradeHistory(currencyPair, start, end string) ([]PoloniexTradeHistory, error) {
	vals := url.Values{}
	vals.Set("currencyPair", currencyPair)

	if start != "" {
		vals.Set("start", start)
	}

	if end != "" {
		vals.Set("end", end)
	}

	resp := []PoloniexTradeHistory{}
	path := fmt.Sprintf("%s/public?command=returnTradeHistory&%s", POLONIEX_API_URL, vals.Encode())
	err := SendHTTPGetRequest(path, true, &resp)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

type PoloniexChartData struct {
	Date            int     `json:"date"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
	Open            float64 `json:"open"`
	Close           float64 `json:"close"`
	Volume          float64 `json:"volume"`
	QuoteVolume     float64 `json:"quoteVolume"`
	WeightedAverage float64 `json:"weightedAverage"`
}

func (p *Poloniex) GetChartData(currencyPair, start, end, period string) ([]PoloniexChartData, error) {
	vals := url.Values{}
	vals.Set("currencyPair", currencyPair)

	if start != "" {
		vals.Set("start", start)
	}

	if end != "" {
		vals.Set("end", end)
	}

	if period != "" {
		vals.Set("period", period)
	}

	resp := []PoloniexChartData{}
	path := fmt.Sprintf("%s/public?command=returnChartData&%s", POLONIEX_API_URL, vals.Encode())
	err := SendHTTPGetRequest(path, true, &resp)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

type PoloniexCurrencies struct {
	Name               string      `json:"name"`
	MaxDailyWithdrawal string      `json:"maxDailyWithdrawal"`
	TxFee              float64     `json:"txFee,string"`
	MinConfirmations   int         `json:"minConf"`
	DepositAddresses   interface{} `json:"depositAddress"`
	Disabled           int         `json:"disabled"`
	Delisted           int         `json:"delisted"`
	Frozen             int         `json:"frozen"`
}

func (p *Poloniex) GetCurrencies() (map[string]PoloniexCurrencies, error) {
	type Response struct {
		Data map[string]PoloniexCurrencies
	}
	resp := Response{}
	path := fmt.Sprintf("%s/public?command=returnCurrencies", POLONIEX_API_URL)
	err := SendHTTPGetRequest(path, true, &resp.Data)

	if err != nil {
		return resp.Data, err
	}
	return resp.Data, nil
}

type PoloniexLoanOrder struct {
	Rate     float64 `json:"rate,string"`
	Amount   float64 `json:"amount,string"`
	RangeMin int     `json:"rangeMin"`
	RangeMax int     `json:"rangeMax"`
}

type PoloniexLoanOrders struct {
	Offers  []PoloniexLoanOrder `json:"offers"`
	Demands []PoloniexLoanOrder `json:"demands"`
}

func (p *Poloniex) GetLoanOrders(currency string) (PoloniexLoanOrders, error) {
	resp := PoloniexLoanOrders{}
	path := fmt.Sprintf("%s/public?command=returnLoanOrders&currency=%s", POLONIEX_API_URL, currency)
	err := SendHTTPGetRequest(path, true, &resp)

	if err != nil {
		return resp, err
	}
	return resp, nil
}

type PoloniexBalance struct {
	Currency map[string]float64
}

func (p *Poloniex) GetBalances() (PoloniexBalance, error) {
	var result interface{}
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_BALANCES, url.Values{}, &result)

	if err != nil {
		return PoloniexBalance{}, err
	}

	data := result.(map[string]interface{})
	balance := PoloniexBalance{}
	balance.Currency = make(map[string]float64)

	for x, y := range data {
		balance.Currency[x], _ = strconv.ParseFloat(y.(string), 64)
	}

	return balance, nil
}

type PoloniexCompleteBalance struct {
	Available float64
	OnOrders  float64
	BTCValue  float64
}

type PoloniexCompleteBalances struct {
	Currency map[string]PoloniexCompleteBalance
}

func (p *Poloniex) GetCompleteBalances() (PoloniexCompleteBalances, error) {
	var result interface{}
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_BALANCES_COMPLETE, url.Values{}, &result)

	if err != nil {
		return PoloniexCompleteBalances{}, err
	}

	data := result.(map[string]interface{})
	balance := PoloniexCompleteBalances{}
	balance.Currency = make(map[string]PoloniexCompleteBalance)

	for x, y := range data {
		dataVals := y.(map[string]interface{})
		balancesData := PoloniexCompleteBalance{}
		balancesData.Available, _ = strconv.ParseFloat(dataVals["available"].(string), 64)
		balancesData.OnOrders, _ = strconv.ParseFloat(dataVals["onOrders"].(string), 64)
		balancesData.BTCValue, _ = strconv.ParseFloat(dataVals["btcValue"].(string), 64)
		balance.Currency[x] = balancesData
	}

	return balance, nil
}

type PoloniexDepositAddresses struct {
	Addresses map[string]string
}

func (p *Poloniex) GetDepositAddresses() (PoloniexDepositAddresses, error) {
	var result interface{}
	addresses := PoloniexDepositAddresses{}
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_DEPOSIT_ADDRESSES, url.Values{}, &result)

	if err != nil {
		return addresses, err
	}

	addresses.Addresses = make(map[string]string)
	data := result.(map[string]interface{})
	for x, y := range data {
		addresses.Addresses[x] = y.(string)
	}

	return addresses, nil
}

func (p *Poloniex) GenerateNewAddress(currency string) (string, error) {
	type Response struct {
		Success  int
		Error    string
		Response string
	}
	resp := Response{}
	values := url.Values{}
	values.Set("currency", currency)

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_GENERATE_NEW_ADDRESS, values, &resp)

	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", errors.New(resp.Error)
	}

	return resp.Response, nil
}

type PoloniexDepositsWithdrawals struct {
	Deposits []struct {
		Currency      string    `json:"currency"`
		Address       string    `json:"address"`
		Amount        float64   `json:"amount,string"`
		Confirmations int       `json:"confirmations"`
		TransactionID string    `json:"txid"`
		Timestamp     time.Time `json:"timestamp"`
		Status        string    `json:"string"`
	} `json:"deposits"`
	Withdrawals []struct {
		WithdrawalNumber int64     `json:"withdrawalNumber"`
		Currency         string    `json:"currency"`
		Address          string    `json:"address"`
		Amount           float64   `json:"amount,string"`
		Confirmations    int       `json:"confirmations"`
		TransactionID    string    `json:"txid"`
		Timestamp        time.Time `json:"timestamp"`
		Status           string    `json:"string"`
		IPAddress        string    `json:"ipAddress"`
	} `json:"withdrawals"`
}

func (p *Poloniex) GetDepositsWithdrawals(start, end string) (PoloniexDepositsWithdrawals, error) {
	resp := PoloniexDepositsWithdrawals{}
	values := url.Values{}

	if start != "" {
		values.Set("start", start)
	} else {
		values.Set("start", "0")
	}

	if end != "" {
		values.Set("end", end)
	} else {
		values.Set("end", strconv.FormatInt(time.Now().Unix(), 10))
	}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_DEPOSITS_WITHDRAWALS, values, &resp)

	if err != nil {
		return resp, err
	}

	return resp, nil
}

type PoloniexOrder struct {
	OrderNumber int64   `json:"orderNumber,string"`
	Type        string  `json:"type"`
	Rate        float64 `json:"rate,string"`
	Amount      float64 `json:"amount,string"`
	Total       float64 `json:"total,string"`
	Date        string  `json:"date"`
	Margin      float64 `json:"margin"`
}

type PoloniexOpenOrdersResponseAll struct {
	Data map[string][]PoloniexOrder
}

type PoloniexOpenOrdersResponse struct {
	Data []PoloniexOrder
}

func (p *Poloniex) GetOpenOrders(currency string) (interface{}, error) {
	values := url.Values{}

	if currency != "" {
		values.Set("currencyPair", currency)
		result := PoloniexOpenOrdersResponse{}
		err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_ORDERS, values, &result.Data)

		if err != nil {
			return result, err
		}

		return result, nil
	} else {
		values.Set("currencyPair", "all")
		result := PoloniexOpenOrdersResponseAll{}
		err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_ORDERS, values, &result.Data)

		if err != nil {
			return result, err
		}

		return result, nil
	}
}

type PoloniexAuthentictedTradeHistory struct {
	GlobalTradeID int64   `json:"globalTradeID"`
	TradeID       int64   `json:"tradeID,string"`
	Date          string  `json:"data,string"`
	Rate          float64 `json:"rate,string"`
	Amount        float64 `json:"amount,string"`
	Total         float64 `json:"total,string"`
	Fee           float64 `json:"fee,string"`
	OrderNumber   int64   `json:"orderNumber,string"`
	Type          string  `json:"type"`
	Category      string  `json:"category"`
}

type PoloniexAuthenticatedTradeHistoryAll struct {
	Data map[string][]PoloniexOrder
}

type PoloniexAuthenticatedTradeHistoryResponse struct {
	Data []PoloniexOrder
}

func (p *Poloniex) GetAuthenticatedTradeHistory(currency, start, end string) (interface{}, error) {
	values := url.Values{}

	if start != "" {
		values.Set("start", start)
	}

	if end != "" {
		values.Set("end", end)
	}

	if currency != "" {
		values.Set("currencyPair", currency)
		result := PoloniexAuthenticatedTradeHistoryResponse{}
		err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_TRADE_HISTORY, values, &result.Data)

		if err != nil {
			return result, err
		}

		return result, nil
	} else {
		values.Set("currencyPair", "all")
		result := PoloniexAuthenticatedTradeHistoryAll{}
		err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_TRADE_HISTORY, values, &result.Data)

		if err != nil {
			return result, err
		}

		return result, nil
	}
}

type PoloniexResultingTrades struct {
	Amount  float64 `json:"amount,string"`
	Date    string  `json:"date"`
	Rate    float64 `json:"rate,string"`
	Total   float64 `json:"total,string"`
	TradeID int64   `json:"tradeID,string"`
	Type    string  `json:"type,string"`
}

type PoloniexOrderResponse struct {
	OrderNumber int64                     `json:"orderNumber,string"`
	Trades      []PoloniexResultingTrades `json:"resultingTrades"`
}

func (p *Poloniex) PlaceOrder(currency string, rate, amount float64, immediate, fillOrKill, buy bool) (PoloniexOrderResponse, error) {
	result := PoloniexOrderResponse{}
	values := url.Values{}

	var orderType string
	if buy {
		orderType = POLONIEX_ORDER_BUY
	} else {
		orderType = POLONIEX_ORDER_SELL
	}

	values.Set("currencyPair", currency)
	values.Set("rate", strconv.FormatFloat(rate, 'f', -1, 64))
	values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	if immediate {
		values.Set("immediateOrCancel", "1")
	}

	if fillOrKill {
		values.Set("fillOrKill", "1")
	}

	err := p.SendAuthenticatedHTTPRequest("POST", orderType, values, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

type PoloniexGenericResponse struct {
	Success int    `json:"success"`
	Error   string `json:"error"`
}

func (p *Poloniex) CancelOrder(orderID int64) (bool, error) {
	result := PoloniexGenericResponse{}
	values := url.Values{}
	values.Set("orderNumber", strconv.FormatInt(orderID, 10))

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_ORDER_CANCEL, values, &result)

	if err != nil {
		return false, err
	}

	if result.Success != 1 {
		return false, errors.New(result.Error)
	}

	return true, nil
}

type PoloniexMoveOrderResponse struct {
	Success     int                                  `json:"success"`
	Error       string                               `json:"error"`
	OrderNumber int64                                `json:"orderNumber,string"`
	Trades      map[string][]PoloniexResultingTrades `json:"resultingTrades"`
}

func (p *Poloniex) MoveOrder(orderID int64, rate, amount float64) (PoloniexMoveOrderResponse, error) {
	result := PoloniexMoveOrderResponse{}
	values := url.Values{}
	values.Set("orderNumber", strconv.FormatInt(orderID, 10))
	values.Set("rate", strconv.FormatFloat(rate, 'f', -1, 64))

	if amount != 0 {
		values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_ORDER_MOVE, values, &result)

	if err != nil {
		return result, err
	}

	if result.Success != 1 {
		return result, errors.New(result.Error)
	}

	return result, nil
}

type PoloniexWithdraw struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func (p *Poloniex) Withdraw(currency, address string, amount float64) (bool, error) {
	result := PoloniexWithdraw{}
	values := url.Values{}

	values.Set("currency", currency)
	values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	values.Set("address", address)

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_WITHDRAW, values, &result)

	if err != nil {
		return false, err
	}

	if result.Error != "" {
		return false, errors.New(result.Error)
	}

	return true, nil
}

type PoloniexFee struct {
	MakerFee        float64 `json:"makerFee,string"`
	TakerFee        float64 `json:"takerFee,string"`
	ThirtyDayVolume float64 `json:"thirtyDayVolume,string"`
	NextTier        float64 `json:"nextTier,string"`
}

func (p *Poloniex) GetFeeInfo() (PoloniexFee, error) {
	result := PoloniexFee{}
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_FEE_INFO, url.Values{}, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (p *Poloniex) GetTradableBalances() (map[string]map[string]float64, error) {
	type Response struct {
		Data map[string]map[string]interface{}
	}
	result := Response{}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_TRADABLE_BALANCES, url.Values{}, &result.Data)

	if err != nil {
		return nil, err
	}

	balances := make(map[string]map[string]float64)

	for x, y := range result.Data {
		balances[x] = make(map[string]float64)
		for z, w := range y {
			balances[x][z], _ = strconv.ParseFloat(w.(string), 64)
		}
	}

	return balances, nil
}

func (p *Poloniex) TransferBalance(currency, from, to string, amount float64) (bool, error) {
	values := url.Values{}
	result := PoloniexGenericResponse{}

	values.Set("currency", currency)
	values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	values.Set("fromAccount", from)
	values.Set("toAccount", to)

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_TRANSFER_BALANCE, values, &result)

	if err != nil {
		return false, err
	}

	if result.Error != "" && result.Success != 1 {
		return false, errors.New(result.Error)
	}

	return true, nil
}

type PoloniexMargin struct {
	TotalValue    float64 `json:"totalValue,string"`
	ProfitLoss    float64 `json:"pl,string"`
	LendingFees   float64 `json:"lendingFees,string"`
	NetValue      float64 `json:"netValue,string"`
	BorrowedValue float64 `json:"totalBorrowedValue,string"`
	CurrentMargin float64 `json:"currentMargin,string"`
}

func (p *Poloniex) GetMarginAccountSummary() (PoloniexMargin, error) {
	result := PoloniexMargin{}
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_MARGIN_ACCOUNT_SUMMARY, url.Values{}, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (p *Poloniex) PlaceMarginOrder(currency string, rate, amount, lendingRate float64, buy bool) (PoloniexOrderResponse, error) {
	result := PoloniexOrderResponse{}
	values := url.Values{}

	var orderType string
	if buy {
		orderType = POLONIEX_MARGIN_BUY
	} else {
		orderType = POLONIEX_MARGIN_SELL
	}

	values.Set("currencyPair", currency)
	values.Set("rate", strconv.FormatFloat(rate, 'f', -1, 64))
	values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	if lendingRate != 0 {
		values.Set("lendingRate", strconv.FormatFloat(lendingRate, 'f', -1, 64))
	}

	err := p.SendAuthenticatedHTTPRequest("POST", orderType, values, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

type PoloniexMarginPosition struct {
	Amount            float64 `json:"amount,string"`
	Total             float64 `json:"total,string"`
	BasePrice         float64 `json:"basePrice,string"`
	LiquidiationPrice float64 `json:"liquidiationPrice"`
	ProfitLoss        float64 `json:"pl,string"`
	LendingFees       float64 `json:"lendingFees,string"`
	Type              string  `json:"type"`
}

func (p *Poloniex) GetMarginPosition(currency string) (interface{}, error) {
	values := url.Values{}

	if currency != "" && currency != "all" {
		values.Set("currencyPair", currency)
		result := PoloniexMarginPosition{}
		err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_MARGIN_POSITION, values, &result)

		if err != nil {
			return result, err
		}

		return result, nil
	} else {
		values.Set("currencyPair", "all")

		type Response struct {
			Data map[string]PoloniexMarginPosition
		}

		result := Response{}
		err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_MARGIN_POSITION, values, &result.Data)

		if err != nil {
			return result, err
		}

		return result, nil
	}
}

func (p *Poloniex) CloseMarginPosition(currency string) (bool, error) {
	values := url.Values{}
	values.Set("currencyPair", currency)
	result := PoloniexGenericResponse{}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_MARGIN_POSITION_CLOSE, values, &result)

	if err != nil {
		return false, err
	}

	if result.Success == 0 {
		return false, errors.New(result.Error)
	}

	return true, nil
}

func (p *Poloniex) CreateLoanOffer(currency string, amount, rate float64, duration int, autoRenew bool) (int64, error) {
	values := url.Values{}
	values.Set("currency", currency)
	values.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	values.Set("duration", strconv.Itoa(duration))

	if autoRenew {
		values.Set("autoRenew", "1")
	} else {
		values.Set("autoRenew", "0")
	}

	values.Set("lendingRate", strconv.FormatFloat(rate, 'f', -1, 64))

	type Response struct {
		Success int    `json:"success"`
		Error   string `json:"error"`
		OrderID int64  `json:"orderID"`
	}

	result := Response{}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_CREATE_LOAN_OFFER, values, &result)

	if err != nil {
		return 0, err
	}

	if result.Success == 0 {
		return 0, errors.New(result.Error)
	}

	return result.OrderID, nil
}

func (p *Poloniex) CancelLoanOffer(orderNumber int64) (bool, error) {
	result := PoloniexGenericResponse{}
	values := url.Values{}
	values.Set("orderID", strconv.FormatInt(orderNumber, 10))

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_CANCEL_LOAN_OFFER, values, &result)

	if err != nil {
		return false, err
	}

	if result.Success == 0 {
		return false, errors.New(result.Error)
	}

	return true, nil
}

type PoloniexLoanOffer struct {
	ID        int64   `json:"id"`
	Rate      float64 `json:"rate,string"`
	Amount    float64 `json:"amount,string"`
	Duration  int     `json:"duration"`
	AutoRenew bool    `json:"autoRenew,int"`
	Date      string  `json:"date"`
}

func (p *Poloniex) GetOpenLoanOffers() (map[string][]PoloniexLoanOffer, error) {
	type Response struct {
		Data map[string][]PoloniexLoanOffer
	}
	result := Response{}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_OPEN_LOAN_OFFERS, url.Values{}, &result.Data)

	if err != nil {
		return nil, err
	}

	if result.Data == nil {
		return nil, errors.New("There are no open loan offers.")
	}

	return result.Data, nil
}

type PoloniexActiveLoans struct {
	Provided []PoloniexLoanOffer `json:"provided"`
	Used     []PoloniexLoanOffer `json:"used"`
}

func (p *Poloniex) GetActiveLoans() (PoloniexActiveLoans, error) {
	result := PoloniexActiveLoans{}
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_ACTIVE_LOANS, url.Values{}, &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (p *Poloniex) ToggleAutoRenew(orderNumber int64) (bool, error) {
	values := url.Values{}
	values.Set("orderNumber", strconv.FormatInt(orderNumber, 10))
	result := PoloniexGenericResponse{}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_AUTO_RENEW, values, &result)

	if err != nil {
		return false, err
	}

	if result.Success == 0 {
		return false, errors.New(result.Error)
	}

	return true, nil
}

func (p *Poloniex) SendAuthenticatedHTTPRequest(method, endpoint string, values url.Values, result interface{}) error {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Key"] = p.AccessKey

	nonce := time.Now().UnixNano()
	nonceStr := strconv.FormatInt(nonce, 10)

	values.Set("nonce", nonceStr)
	values.Set("command", endpoint)

	hmac := GetHMAC(HASH_SHA512, []byte(values.Encode()), []byte(p.SecretKey))
	headers["Sign"] = HexEncodeToString(hmac)

	path := fmt.Sprintf("%s/%s", POLONIEX_API_URL, POLONIEX_API_TRADING_ENDPOINT)
	resp, err := SendHTTPRequest(method, path, headers, bytes.NewBufferString(values.Encode()))

	if err != nil {
		return err
	}

	err = JSONDecode([]byte(resp), &result)

	if err != nil {
		return errors.New("Unable to JSON Unmarshal response.")
	}
	return nil
}
