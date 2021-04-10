package main

import (
	"flag"
	"log"
	"math"
	"os"

	"github.com/joho/godotenv"
	"github.com/tacotokyo/bybit-auto-compound/api"
)

func main() {
	envfile := flag.String("envfile", "", "env file path (e.g. /path/to/.env)")
	coin := flag.String("coin", "", "target coin (e.g. BTC, ETH)")
	flag.Parse()

	if *envfile == "" || *coin == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	err := godotenv.Load(*envfile)
	if err != nil {
		log.Fatalf("%v", err)
	}
	key, secret := keysecret()
	bybit := api.New(key, secret)
	start(bybit, *coin)
}

func start(bybit *api.BybitApi, coin string) {
	log.Printf("cancel all orders: %s", coin)
	err := bybit.CancelAll(coin)
	if err != nil {
		log.Fatalf("%v", err)
	}
	pos, err := bybit.Positions(coin)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if pos.Side == "Buy" {
		log.Fatalln("currently position is long. cannot make short position")
	}
	if pos.Side == "Sell" && pos.EffectiveLeverage > 1 {
		log.Fatalln("current position is over x1 leverage")
	}
	if pos.Side == "Sell" && pos.Leverage > 1 {
		err = bybit.LeverageX1(coin)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	bal, err := bybit.AvailableBalance(coin)
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("%s available balance: %v", coin, bal)
	ask, bid, err := bybit.Price(coin)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if bal*bid > 1 {
		size := int64(math.Floor(ask * bal))
		log.Printf("order: %s short price: %v size: %v", coin, ask, size)
		id, err := bybit.Sell(coin, ask, size)
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("ordered: %s", id)
	} else {
		log.Printf("no available balance. skip order")
	}
}

func keysecret() (key, secret string) {
	key = os.Getenv("MAIN_BYBIT_KEY")
	secret = os.Getenv("MAIN_BYBIT_SECRET")
	if key == "" && secret == "" {
		log.Fatalln("key/secret not found")
	}
	return
}
