package main

import (
	"log"
	"math"
	"os"

	"github.com/joho/godotenv"
	"github.com/tacotokyo/bybit-auto-compound/api"
)

func main() {
	godotenv.Load()
	key, secret := keysecret()
	bybit := api.New(key, secret)
	coin := "xrp"
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
		bybit.Sell(coin, ask, size)
	} else {
		log.Printf("no available balance. skip order")
	}
}

func keysecret() (key, secret string) {
	godotenv.Load()
	key = os.Getenv("MAIN_BYBIT_KEY")
	secret = os.Getenv("MAIN_BYBIT_SECRET")
	if key == "" && secret == "" {
		log.Fatalln("key/secret not found")
	}
	return
}
