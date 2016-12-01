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
	POLONIEX_ORDER_TRADES           = "returnOrderTrades"
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

	fee = .0015
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
	_ = currency

	toTrade := "ETH"
	log.Printf("balance: %s %f", toTrade, p.balance(toTrade))

	//	for p.Enabled {

	//go func() {
	//log.Printf("Poloniex=%s Last=%f High=%f Low=%f Volume=%f Ask=%f Bid=%f EMA13=%f EMA41=%f\n", currency, t.Last, t.High24Hr, t.Low24Hr, t.QuoteVolume, t.LowestAsk, t.HighestBid, f, s)
	//}()

	// TODO add slippage penalty to sims
	// TODO add a stop limit for -20% for every order? just in case...
	// TODO this fucked up, test 2016/06/13 00:00:18 WARN couldn't place order. bailing for now. maybe postOnly=true? currency=BTC_ETH rate=0.023201 amount=1.897850 lending_rate=0.005000 buy=false err=error unmarshaling json: json: cannot unmarshal string into Go value of type int64 text: {"success":1,"message":"Margin order placed.","orderNumber":"67936208935","resultingTrades":[{"amount":"1.89785002","date":"2016-06-13 04:00:18","rate":"0.02320394","total":"0.04403759","tradeID":"11160991","type":"sell"}]}
	// TODO last candle / tick candles seem fucked up. need to wait until past candle close then get last candle close, not current candle

	p.realTrade(toTrade, currency)

	// BUY THE FARM
	//p.allIn(toTrade, currency, true)
	//p.CloseMarginPosition(currency)

	// SELL THE FARM
	//p.allIn(toTrade, currency, false)
	//p.CloseMarginPosition(currency)

	// TODO sims lay below, extract somehow
	//	p.tryAll(currency)

	// TODO test:
	// #1 22/31/2 @120 .01 breakout -> 1200@6mos .59 tharp 90 trades || .001 breakout -> 1350@6mos .82 tharp 114 trades -> non-cum: profit%=265.179925 profit=2.651799 fees=0.178200 %win=0.392857 avgW=0.101253 %loss=0.607143 avgL=-0.026520 trades=112 tharp=0.892796

	// #2 19/23/5 @120 .01 breakout -> 1500%@6mos .63 tharp 93 trades -> non-cum: profit%=296.948102 profit=2.969481 fees=0.150100 %win=0.484211 avgW=0.095179 %loss=0.515789 avgL=-0.028750 trades=95 tharp=1.087239
	// #2 14/24/5 @120 .01 breakout -> 1500%@6mos .59 tharp 106 trades -> non-cum: profit%=296.612907 profit=2.966129 fees=0.168000 %win=0.452830 avgW=0.096954 %loss=0.547170 avgL=-0.029098 trades=106 tharp=0.961669

	//for _, days := range []int{7, 14, 21, 30, 60, 90, 120, 150, 180, 210, 240} {
	//for _, days := range []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200} {
	//	log.Printf("days: %d", days)
	//p.tryOne(currency, days, 5, 1, 5)
	//p.tryOne(currency, days, 29, 37, 120)
	//p.tryOne(currency, days, 17, 30, 120)
	//p.tryOne(currency, days, 22, 31, 120) // sig=2
	//p.tryOne(currency, days, 47, 81, 120)
	//p.tryOne(currency, days, 40, 86, 120)
	//	p.tryOne(currency, days, 50, 85, 120)
	//p.tryOne(currency, days, 37, 47, 120)
	//p.tryOne(currency, days, 33, 63, 120)
	//p.tryOne(currency, days, 46, 49, 120)
	//p.tryOne(currency, days, 19, 23, 120) // sig=5
	//p.tryOne(currency, days, 14, 24, 120) // sig=5
	//p.tryOne(currency, days, 11, 29, 120) // sig=5 SUPREME LEADA
	//p.tryOne(currency, days, 45, 200, 120)
	//p.tryOne(currency, days, 50, 175, 120) // sig=10
	//p.tryOne(currency, days, 40, 200, 120) // sig=10
	//p.tryOne(currency, days, 3, 8, 120)    // sig=2 breakout=true
	//p.tryOne(currency, days, 1, 40, 120)   // sig=5 breakout=true
	//p.tryOne(currency, days, 2, 14, 120)   // sig=2 breakout=true
	//p.tryOne(currency, days, 2, 45, 120)   // sig=2 breakout=true
	//p.tryOne(currency, days, 23, 115, 120) // sig=10 breakout=true
	//p.tryOne(currency, days, 12, 19, 120)  // sig=10 breakout=true
	//}

	//time.Sleep(time.Second * p.RESTPollingDelay)
	//}
}

func (p *Poloniex) balance(side string) float64 {
	acc, err := p.GetAvailableBalances()
	if err != nil {
		log.Fatalf("couldn't get margin account info: %v", err)
	}

	return acc["margin"][side]
}

func (p *Poloniex) allIn(side, currency string, buy bool) {
	bal := p.balance(side)
	log.Printf("account balance %s: %f", side, bal)

	// BUY
	p.trade(currency, bal, buy)
}

func (p *Poloniex) realTrade(side, currency string) {
	// TODO close open orders when this shuts down?

	// TODO close open orders before trading? can fix later..

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

	// NOTE: if we happen to have an open position, let back data determine whether we should keep it open

	log.Printf("open pos found: type=%s price=%f amount=%f total=%f p/l=%f lending_fees=%f", open.Type, open.BasePrice, open.Amount, open.Total, open.ProfitLoss, open.LendingFees)

	var profit, fees float64
	fast, slow, sig := 50, 85, 5 // TODO make these configuramable ?
	const candle = 7200          // 2hr candle; candlestick period in seconds; valid values are 300, 900, 1800, 7200, 14400, and 86400
	emaFast := ema(fast)
	emaSlow := ema(slow)
	signal := ema(sig)

	// get enough back data to seed the fast / slow emas so we can start trading..
	// get 1 more than the last tick and save a tick so that we might immediately
	// close/open a position. the loop ends up always getting the tick from approx:
	// now() - candle, in order to read the full candle close (for charts yo)

	// TODO need to get candle ~1 minute after it closes, so data is complete. meh

	lastData := time.Now().Add(-2 * candle * time.Second)
	timemachine := time.Duration(sig+int(math.Max(float64(fast), float64(slow)))) * candle * time.Second
	start := strconv.Itoa(int(lastData.Add(-timemachine).Unix()))
	end := strconv.Itoa(int(lastData.Unix()))
	period := strconv.Itoa(candle)
	chart, err := p.GetChartData(currency, start, end, period)
	if err != nil {
		log.Fatalf("issue going back in time; check the flux capacitor. err:", err)
	}

	// initialize all the things
	for _, pt := range chart {
		log.Printf("backdata t=%v high=%f low=%f open=%f close=%f %%=%f volume=%f", time.Unix(int64(pt.Date), 0), pt.High, pt.Low, pt.Open, pt.Close, 100*((pt.Close-pt.Open)/pt.Open), pt.Volume)
		tradeMACD(pt.Close, &lastBuy, &profit, &fees, &pos, emaFast, emaSlow, signal)
		lastData = time.Unix(int64(pt.Date), 0)
	}

	// TODO this doesn't gracefully handle missing ticks but trying to use lastData..lastData+candle kept giving lastData pt so fuck it
	lastCandle := func(lastData time.Time) PoloniexChartData {
		// retry for 2 minutes, and then bail
		var errCount int
		for ; ; time.Sleep(4 * time.Second) {
			tryS := strconv.Itoa(int(lastData.Add(candle * time.Second).Unix()))
			tryE := strconv.Itoa(int(lastData.Add(2 * candle * time.Second).Unix()))
			// just get one data point
			c, err := p.GetChartData(currency, tryS, tryE, period)
			if err != nil {
				errCount++
				if errCount > 30 { // TODO assert that sleep * this is < fast ema
					log.Fatalf("couldn't get price, maybe polo is down? err:", err)
				}
				continue
			}

			if len(c) < 1 || c[0].Date == 0 { // invalid, try again.. it takes a few tries to get next candle
				continue
			}

			pt := c[0]
			for _, p := range c { // sometimes we get 2 data points, probably should check these are contiguous but blech
				if time.Unix(int64(p.Date), 0).Unix() != lastData.Unix() {
					pt = p
					break
				}
			}

			if time.Unix(int64(pt.Date), 0).Unix() == lastData.Unix() {
				continue
			}

			return pt
		}
	}

	// immediately do the first tick so that we might open a position
	tick := time.After(0)
	for {
		<-tick

		pt := lastCandle(lastData)

		lastData = time.Unix(int64(pt.Date), 0)

		nextTime := lastData.Add(2 * candle * time.Second).Sub(time.Now())
		tick = time.After(nextTime)

		log.Printf("tick t=%v high=%f low=%f open=%f close=%f %%=%f volume=%f pos=%s@%f", time.Unix(int64(pt.Date), 0), pt.High, pt.Low, pt.Open, pt.Close, 100*((pt.Close-pt.Open)/pt.Open), pt.Volume, pos, lastBuy)

		last := pos
		beforeProfit := profit
		tradeMACD(pt.Close, &lastBuy, &profit, &fees, &pos, emaFast, emaSlow, signal)

		// execute trade if our position changed (but not our first time determining direction)
		if last != pos && !(lastBuy == 0 && last == none) {
			log.Printf("profits: total=%f total%%=%f last=%f last%%=%f", profit, 100*(profit/1), profit-beforeProfit, 100*((profit-beforeProfit)/beforeProfit))

			_, err = p.CloseMarginPosition(currency)
			if err != nil {
				log.Printf("WARN couldn't close margin position, maybe there isn't one? err=%v", err)
			}

			// close our previous order and then go all in, so that we invest any earnings
			// (and don't take margin if we lost dollas and go short)

			switch pos {
			case none:
			case long, short:
				p.allIn(side, currency, pos == long)
			}
		}
	}
}

// TODO add retries to this so we don't miss a move
func (p *Poloniex) trade(currency string, amount float64, buy bool) {
	str := "LONG"
	if !buy {
		str = "SHORT"
	}

	// try to be a maker and save on fees. we're already guessing direction so adding
	// an order slightly in front of it, if it doesn't hit we probably don't want to be
	// in that position anyway.

	// TODO test that flipping long/short closes other order

	var maxLendingRate float64 = .005
	// TODO calculate maxRate later
	//if !buy {
	//loansA, err := p.GetOpenLoanOffers()
	//if err != nil {
	//log.Printf("WARN couldn't get loan offers. currency=%s err=%v", currency, err)
	//return
	//}
	//var filled float64
	//for _, l := range loansA[currency] {
	//maxLendingRate = l.Rate
	//filled += amount
	//if filled >= amount {
	//break
	//}
	//}

	//if filled < amount {
	//log.Printf("WARN not enough loans open to fill short order, this could mean high fees")
	//}

	//// add 10% to make sure it fills
	//const maxTolerableRate = .005 // .5% -- yikes
	//if maxLendingRate+(maxLendingRate*.1) > maxTolerableRate {
	//log.Printf("WARN not making trade because lending fees are above tolerable rates. rate=%f tolerable=%f", maxLendingRate, maxTolerableRate)
	//return
	//}
	//}

	bestPrice := func() (float64, error) {
		tickA, err := p.GetTicker()
		if err != nil {
			log.Printf("WARN couldn't get ticker. currency=%s err=%v", currency, err)
			return 0, err
		}

		tick := tickA[currency]

		// try to take the maker fee, but not very hard so we don't wait...
		var rate float64
		const outFront = .00001 // .01%
		if buy {
			rate = tick.LowestAsk - (outFront * tick.LowestAsk)
		} else {
			rate = tick.HighestBid + (outFront * tick.HighestBid)
		}
		return rate, nil
	}

	// TODO jesus fuck clean this up

	// TODO we could split the order across N orders and try to move the market a little ;)
	rate, err := bestPrice()
	if err != nil {
		return
	}

	startRate := rate
	var filled, avg float64
	var trades uint64

	// shave fees so we don't have to borrow anything
	amount *= .975

	log.Printf("opening trade %s %s %f@%f", currency, str, amount, rate)
	postOnly := false // only take maker fee, TODO do this later? if it starts moving fast, would prefer to just get in so...
	order, err := p.PlaceMarginOrder(currency, rate, amount, maxLendingRate, postOnly, buy)
	if err != nil {
		log.Printf("WARN couldn't place order. bailing for now. maybe postOnly=true? currency=%s rate=%f amount=%f lending_rate=%f buy=%v err=%v", currency, rate, amount, maxLendingRate, buy, err)
		return
	}

	tradeIDs := make(map[int64]struct{}) // don't count trades twice
	orderNum := order.OrderNumber

	// avg = (amount1 * rate1) + (amount2 * rate2) + ...
	//       ------------------------------------
	//              amount1 + amount2 + ...

	for _, t := range order.Trades {
		tradeIDs[t.TradeID] = struct{}{}
		trades++
		filled += t.Amount
		avg += t.Total // amount * rate
		log.Printf("trade executed for order %d: amount=%f rate=%f total=%f type=%s time=%s originalAmount=%f left=%f filled=%f %%filled=%f", orderNum, t.Amount, t.Rate, t.Total, t.Type, t.Date, amount, amount-filled, filled, 100*(filled/amount))
	}

	if amount-filled <= 0.00000 { // fuckin floats
		avg /= amount
		log.Printf("filled order for amount %f. trades=%d firstRate=%f avgRate=%f", amount, trades, startRate, avg)
		return
	}

	// check every 5s to see if our order filled, after 1m go change our order
	start := time.Now()
	for ; ; time.Sleep(5 * time.Second) {

		// if it's been a minute, cancel our order first, then check trades, and then
		// if we still need to place another order to fulfill the amount, do so.
		if time.Since(start) > 1*time.Minute {
			// NOTE: MoveOrder could partially fill while we're doing calculation, so
			// we need to cancel and then put it another order.
			log.Printf("canceling open order, will make new one if still not filled. order=%d", orderNum)
			_, err := p.CancelOrder(orderNum)
			if err != nil {
				log.Printf("couldn't cancel order. maybe filled? err=%v order=%d originalAmount=%f left=%f filled=%f %%filled=%f", err, orderNum, amount, amount-filled, filled, 100*(filled/amount))
			}
		}

		orderTrades, _ := p.GetOrderTrades(orderNum)
		// ignore errors, we don't care if no trades were executed for this order (that's the point)

		for _, t := range orderTrades {
			if _, ok := tradeIDs[t.TradeID]; ok {
				continue // don't count trades twice
			}

			tradeIDs[t.TradeID] = struct{}{}
			trades++
			filled += t.Amount
			avg += t.Total // (amount * rate)
			log.Printf("trade executed for order %d: amount=%f rate=%f total=%f type=%s time=%s originalAmount=%f left=%f filled=%f %%filled=%f", orderNum, t.Amount, t.Rate, t.Total, t.Type, t.Date, amount, amount-filled, filled, 100*(filled/amount))
			start = time.Now() // update this so that we sit at this price for a minute longer since we filled something
		}

		if amount-filled <= 0.00000 { // fuckin floats
			avg /= amount
			log.Printf("filled order for amount %f. trades=%d firstRate=%f avgRate=%f", amount, trades, startRate, avg)
			return
		}

		// at this point, we cancelled the order, and we haven't yet filled it, so place another
		if time.Since(start) > 1*time.Minute {
			rate, err = bestPrice()
			if err != nil {
				return
			}

			log.Printf("re-opening trade at different price %s %s %f@%f", currency, str, amount-filled, rate)
			order, err = p.PlaceMarginOrder(currency, rate, amount-filled, maxLendingRate, postOnly, buy)
			if err != nil {
				log.Printf("WARN couldn't place order. bailing. maybe postOnly=true? currency=%s rate=%f amount=%f lending_rate=%f buy=%v err=%v", currency, rate, amount, maxLendingRate, buy, err)
				return
			}

			orderNum = order.OrderNumber
			start = time.Now()

			for _, t := range order.Trades {
				tradeIDs[t.TradeID] = struct{}{}
				trades++
				filled += t.Amount
				avg += t.Total // amount * rate
				log.Printf("trade executed for order %d: amount=%f rate=%f total=%f type=%s time=%s originalAmount=%f left=%f filled=%f %%filled=%f", orderNum, t.Amount, t.Rate, t.Total, t.Type, t.Date, amount, amount-filled, filled, 100*(filled/amount))
			}

			if amount-filled <= 0.00000 { // fuckin floats
				avg /= amount
				log.Printf("filled order for amount %f. trades=%d firstRate=%f avgRate=%f", amount, trades, startRate, avg)
				return
			}
		}
	}
}

func (p *Poloniex) tryOne(currency string, days, fast, slow, tick int) {
	sig := 5
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

	tharp, profit, f := tryEma(fast, slow, sig, tick, c)

	log.Printf("%s profit: f= %d s= %d t= %d profit%%= %f profit= %f fees= %f price= %f tharp= %f", currency, fast, slow, tick*5, 100*(profit/1), profit, f, first, tharp)

	// TODO print out each trade
}

func (p *Poloniex) tryAll(currency string) {
	var maxProfit, fees, first, bestTh float64
	var bestF, bestS, bestSig int

	const maxFast = 50
	const maxSlow = 200
	//const maxTick = 24
	const maxSig = 10
	// sig := 5
	tick := 24 // 2hr candle

	// y = tick, x = fast/slow ema combos
	matrix := make([][]float64, maxFast*maxSlow)

	var cur, avg [maxFast * maxSlow * maxSig]rank

	var left [][3]int
	for fast := 1; fast <= maxFast; fast++ {
		for slow := 1; slow <= maxSlow; slow++ {
			left = append(left, [3]int{fast, slow, 0})
			for sig := 1; sig <= maxSig; sig++ {
				p := [3]int{fast, slow, sig}
				avg[(sig-1)+(slow-1)*maxSig+(fast-1)*maxSlow*maxSig] = rank{p: p}
			}
		}
	}

	trials := 210
	shortest := 210
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
				matrix[i] = make([]float64, maxSig)
				for sig := 1; sig <= maxSig; sig++ { // up to 2 hours
					tharp, profit, f := tryEma(fast, slow, sig, tick, c)
					p := 100 * (profit / 1)
					matrix[i][sig-1] = p
					if profit > maxProfit {
						//if tharp > bestTh && !math.IsNaN(tharp) {
						maxProfit, fees, bestF, bestS, bestSig, bestTh = profit, f, fast, slow, sig, tharp
					}
					cur[(sig-1)+(slow-1)*maxSig+(fast-1)*maxSlow*maxSig] = rank{p: [3]int{fast, slow, sig}, a: p}
				}
				i++
			}
		}

		// sort by profits, highest profit first (= lowest rank = best)
		sort.Stable(profits(cur[:]))
		for i, c := range cur {
			// track total ranks thus far, average later
			avg[(c.p[2]-1)+(c.p[1]-1)*maxSig+(c.p[0]-1)*maxSlow*maxSig].a += float64(i + 1)
		}
	}

	for i := range avg {
		avg[i].a /= float64(trials - shortest + 1)
	}

	sort.Stable(ranks(avg[:]))

	fmt.Println("leaderboard: [fast/slow/sig]: [avg rank]")
	fmt.Println()
	for i, a := range avg {
		fmt.Printf("%10d: %3d/%3d/%3d: %9.3f\n", i+1, a.p[0], a.p[1], a.p[2], a.a)
	}

	//fast, slow, tick := 13, 41, 12
	//profit, f := tryEma(fast, slow, tick, c)
	//maxProfit, fees, bestF, bestS, bestT = profit, f, fast, slow, tick
	log.Printf("%s best: f= %d s= %d sig= %d t= %d profit%%= %f profit= %f fees= %f price= %f tharp= %f", currency, bestF, bestS, bestSig, tick, 100*(maxProfit/1), maxProfit, fees, first, bestTh)

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
	// graph(matrix, left, bestTh)
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
			if math.IsNaN(p) {
				p = 0 // :(
			}
			r, g, b := color(p, max*.1, max*.5)
			fmt.Printf("%s ", rgbterm.String(fmt.Sprintf("%-6.3f", p), r, g, b, 0, 0, 0)) // Multiply by some constant to make it human readable
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

func tryEma(fast, slow, sig, tickC int, lines []PoloniexChartData) (tharp, profit, fees float64) {
	emaFast := ema(fast)
	emaSlow := ema(slow)
	signal := ema(sig)
	_ = signal
	var lastDir dir
	var tick int
	var lastBuy float64

	var tWin, tLoss float64
	var winners, losers []float64

	for _, l := range lines {
		tick++

		// profit calc from log of prices
		if tick%tickC == 0 {
			oldP := profit
			tradeMACD(l.Close, &lastBuy, &profit, &fees, &lastDir, emaFast, emaSlow, signal)
			// tradeEMA(l.Close, &lastBuy, &profit, &fees, &lastDir, emaFast, emaSlow)

			if oldP > profit {
				p := profit - oldP
				tLoss += p
				losers = append(losers, p)
			} else if oldP < profit {
				p := profit - oldP
				tWin += p
				winners = append(winners, p)
			}
		}
	}

	// TODO Tharp expectancy = (average $ winners * win % + average $ losers * lose %) / (−average $ losers)

	winPercent := float64(len(winners)) / float64(len(winners)+len(losers))
	lossPercent := float64(len(losers)) / float64(len(winners)+len(losers))
	avgWin := tWin / float64(len(winners))
	avgLoss := tLoss / float64(len(losers))

	tharp = (avgWin*winPercent + avgLoss*lossPercent) / -avgLoss

	log.Printf("expectancy: f= %d s= %d t= %d sig= %d profit%%= %f profit= %f fees= %f %%win= %f avgW= %f %%loss= %f avgL= %f trades= %d tharp= %f", fast, slow, tickC*5, sig, 100*(profit/1), profit, fees, winPercent, avgWin, lossPercent, avgLoss, len(winners)+len(losers), tharp)

	return tharp, profit, fees
}

func tradeMACD(price float64, lastBuy, profit, fees *float64, last *dir, emaFast, emaSlow, signal func(float64) float64) {
	f := emaFast(price)
	s := emaSlow(price)

	if f == notTrained || s == notTrained {
		return
	}

	// MACD Line: (12-day EMA - 26-day EMA)
	// Signal Line: 9-day EMA of MACD Line
	macd := f - s
	v := signal(macd)

	if v == notTrained {
		return
	}

	// stop losses
	//stopLoss := .05
	// TODO add drift to these, also they aren't precise since w/i the candle
	// they could have been tipped off, but it will help mitigate larger losses in data regardless
	//stopShortPrice := (*lastBuy) + (stopLoss * (*lastBuy))
	//stopLongPrice := (*lastBuy) - (stopLoss * (*lastBuy))
	//stopShort := price >= stopShortPrice
	//stopLong := price <= stopLongPrice

	// TODO figure these out..
	//stopShort, stopLong := false, false

	// TODO make sure macd actually 'breaks out' before flipping
	//diff := (math.Abs(macd-v) / ((macd + v) / 2))
	//breakout := true // diff >= .0000001 // make sure breakout is real to avoid fakeouts. TODO why is this fucked?

	// go long if macd > 0 && macd > v
	// close long if macd < 0 || macd < v
	// go short if macd < 0 && macd < v
	// close short if macd > 0 || macd > v

	// TODO calculate exposure

	// close order first
	if *lastBuy > 0 && *last == long && ( /*macd < 0 ||*/ macd < v) {
		// close long
		mult := 1. + *profit // NOTE: add profit back for compound

		var p float64
		p = mult * ((price - *lastBuy) / *lastBuy)
		// NOTE compound ends up weighting later profits higher, which sucks (but shiny)

		// log.Printf("msg=LONGPROFITS buy=%f price=%f profit=%f gross_profit=%f net_profit=%f fee=%f", *lastBuy, price, p, *profit+p, *profit+p-f, f)
		f := (fee * mult) + (.0002 * mult)
		*fees += f
		*profit += p - f
		*last = none
	} else if *lastBuy > 0 && *last == short && ( /*macd > 0 ||*/ macd > v) {
		// close short
		mult := 1. + *profit // NOTE: add profit back for compound

		var p float64
		p = mult * ((*lastBuy - price) / *lastBuy)
		// change profit in terms of eth_btc to be in terms of eth

		// log.Printf("msg=SHORTPROFITS buy=%f price=%f profit=%f gross_profit=%f net_profit=%f fee=%f", *lastBuy, price, p, *profit+p, *profit+p-f, f)
		//const lending = .0002
		f := (fee * mult) + (.0002 * mult)
		*fees += f
		*profit += p - f
		*last = none
	}

	// open new ones, if necessary
	if /*macd < 0 &&*/ macd < v { // TODO confirm < 0 ?
		// only go short if on the first trade, we were looking for a short xover or we were in cash.
		// i.e. don't make the first trade until the first xover...
		if *last == none || (*lastBuy == 0 && *last == long) {
			*last = short
			*lastBuy = price
			//	log.Printf("msg=SHORT price=%f", price)
		}
	} else if /*macd > 0 &&*/ macd > v { // TODO confirm > 0 ?
		if *last == none || (*lastBuy == 0 && *last == short) {
			*last = long
			*lastBuy = price
			//	log.Printf("msg=LONG price=%f", price)
		}
	}
}

func tradeEMA(price float64, lastBuy, profit, fees *float64, last *dir, emaFast, emaSlow func(float64) float64) {
	f := emaFast(price)
	s := emaSlow(price)
	if f == notTrained || s == notTrained {
		return
	}

	if *last == none && *lastBuy == 0 { // set so we can see direction
		if f > s {
			*last = long
		} else {
			*last = short
		}
		return
	}

	if f < s && *last == long {
		*last = short

		if *lastBuy > 0 {
			mult := 1. //+ *profit // NOTE: add profit back for compound
			p := mult * ((price - *lastBuy) / *lastBuy)

			// log.Printf("msg=SHORT msg2=LONGPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
			f := (fee * mult)

			*fees += f
			*profit += p - f
		}
		*lastBuy = price
	} else if f > s && *last == short {
		*last = long

		if *lastBuy > 0 {
			mult := 1. //+ *profit // NOTE: add profit back for compound
			p := mult * ((*lastBuy - price) / *lastBuy)
			// log.Printf("msg=LONG msg2=SHORTPROFITS buy=%f price=%f profit=%f total_profit=%f", *lastBuy, price, p, *profit+p)
			f := (fee * mult) + (.0002 * mult)

			*fees += f
			*profit += p - f
		}
		*lastBuy = price
	}
}

type dir uint8

const (
	none dir = iota
	short
	long

	notTrained = -1
)

func (d dir) String() string {
	switch d {
	case none:
		return "none"
	case short:
		return "short"
	case long:
		return "long"
	default:
		panic("invalid direction")
	}
}

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

func (p *Poloniex) GetCompleteBalances(typ string) (PoloniexCompleteBalances, error) {
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

func (p *Poloniex) GetOrderTrades(orderID int64) ([]PoloniexOrderTrades, error) {
	values := url.Values{}
	values.Set("orderNumber", strconv.FormatInt(orderID, 10))

	var result []PoloniexOrderTrades
	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_ORDER_TRADES, values, &result)

	return result, err
}

// TODO no fees? :(
type PoloniexResultingTrades struct {
	Amount  float64 `json:"amount,string"`
	Date    string  `json:"date"`
	Rate    float64 `json:"rate,string"`
	Total   float64 `json:"total,string"`
	TradeID int64   `json:"tradeID,string"`
	Type    string  `json:"type"`
}

// since the api is all fucked up and sometimes tradeids aren't strings..
type PoloniexOrderTrades struct {
	Amount  float64 `json:"amount,string"`
	Date    string  `json:"date"`
	Rate    float64 `json:"rate,string"`
	Total   float64 `json:"total,string"`
	TradeID int64   `json:"tradeID"`
	Type    string  `json:"type"`
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

func (p *Poloniex) GetAvailableBalances() (map[string]map[string]float64, error) {
	type Response struct {
		Data map[string]map[string]interface{}
	}
	result := Response{}

	err := p.SendAuthenticatedHTTPRequest("POST", POLONIEX_AVAILABLE_BALANCES, url.Values{}, &result.Data)

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

func (p *Poloniex) PlaceMarginOrder(currency string, rate, amount, lendingRate float64, postOnly, buy bool) (PoloniexOrderResponse, error) {
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

	if postOnly {
		values.Set("postOnly", "1")
	}

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
		return err
	}
	return nil
}
