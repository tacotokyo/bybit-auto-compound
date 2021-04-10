package api

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestAvailableBalance(t *testing.T) {
	key, secret := keysecret(t)
	api := New(key, secret)
	bal, err := api.AvailableBalance("btc")
	log.Printf("%v %v", bal, err)
}

func TestPosition(t *testing.T) {
	key, secret := keysecret(t)
	api := New(key, secret)
	pos, err := api.Positions("btc")
	log.Printf("%#v %v", pos, err)
}

func TestCancellAll(t *testing.T) {
	key, secret := keysecret(t)
	api := New(key, secret)
	err := api.CancelAll("btc")
	log.Printf("%v", err)
}

func TestSell(t *testing.T) {
	key, secret := keysecret(t)
	api := New(key, secret)
	id, err := api.Sell("btc", 70000, 1)
	log.Printf("%v %v", id, err)
}

func keysecret(t *testing.T) (key, secret string) {
	godotenv.Load()
	key = os.Getenv("TEST_BYBIT_KEY")
	secret = os.Getenv("TEST_BYBIT_SECRET")
	if key == "" && secret == "" {
		t.Skip("no key/secret specified in env file for testing")
	}
	return
}
