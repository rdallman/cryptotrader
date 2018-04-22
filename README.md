## FORKED notice

this is a heavily modified version of the forked which needs to be better
encapsulated should I one day regain motivation to unearth this project.
however, I believe in OSS and there are some what I believe to be useful bits
I'd like to keep for eternity and share with other people so that they can
make better things too. I am not responsible for any monetary wins or losses,
but you are free to use this code. I can not honestly recommend willy nilly
pointing it at any of your accounts unless you have run sufficient backtesting
and really know what you are doing.

modifications include:
* backtesting support of algorithms (have to modify & recompile, atm), with
  [matrix] heatmaps of profitability and some computations to determine
  highest profitability 'hot spots' to avoid anomolous curve fitting, as well
  as tharp expectancy and fees accrued.
* ability to monitor and place trades based on heuristics
* simple ema / macd crossover trade & backtesting support (short & long, on margin)

this is pretty gnarly and not easy to modify, written in a coffee fueled
greedy rage, but there's useful stuff in there. one day, i'll get back to
this...

## Cryptocurrency trading bot written in Golang

[![Build Status](https://travis-ci.org/thrasher-/gocryptotrader.svg?branch=master)](https://travis-ci.org/thrasher-/gocryptotrader)
[![Test Coverage](https://codecov.io/github/thrasher-/gocryptotrader/coverage.svg?branch=master)](https://codecov.io/github/thrasher-/gocryptotrader?branch=master)

A cryptocurrency trading bot supporting multiple exchanges written in Golang. 

**Please note that this bot is under development and is not ready for production!**

## Exchange Support Table

| Exchange | REST API | Streaming API | FIX API |
|----------|------|-----------|-----|
| Alphapoint | Yes  | Yes        | NA  |
| ANXPRO | Yes  | No        | NA  |
| Bitfinex | Yes  | Yes        | NA  |
| Bitstamp | Yes  | Yes       | NA  |
| BTCC | Yes  | Yes     | No  |
| BTCE     | Yes  | NA        | NA  |
| BTCMarkets | Yes | NA       | NA  |
| Coinbase | Yes | Yes | No|
| Gemini | Yes | NA | NA |
| Huobi | Yes | Yes |No |
| ItBit | Yes | NA | NA |
| Kraken | Yes | NA | NA |
| LakeBTC | Yes | Yes | NA | 
| LocalBitcoins | Yes | NA | NA |
| OKCoin (both) | Yes | Yes | No |
| Poloniex | Yes | Yes | NA |

** NA means not applicable as the Exchange does not support the feature.

## Current Features
+ Support for all Exchange fiat and digital currencies, with the ability to individually toggle them on/off.
+ REST API support for all exchanges.
+ Websocket support for applicable exchanges.
+ Ability to turn off/on certain exchanges.
+ Ability to adjust manual polling timer for exchanges.
+ SMS notification support via SMS Gateway.
+ Basic event trigger system.

## Planned Features
+ WebGUI.
+ FIX support.
+ Expanding event trigger system.
+ TALib.
+ Trade history summary generation for tax purposes.

Please feel free to submit any pull requests or suggest any desired features to be added.

## Compiling instructions
Download Go from https://golang.org/dl/  
Using a terminal, type go get github.com/thrasher-/gocryptotrader  
Change directory to the package directory, then type go install.  
Copy config_example.json to config.json.  
Make any neccessary changes to the config file.  
Run the application!  

## Binaries
Binaries will be published once the codebase reaches a stable condition.

