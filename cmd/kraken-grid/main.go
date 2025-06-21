package main

import (
	"fmt"
	"log"
	"os"

	"github.com/krakenfx/api-go/pkg/spot"
	"github.com/spf13/viper"
)

type Configuration struct {
	Positions []Position
}

type Position struct {
	Name      string `mapstructure:"name"`
	BuyLimit  Limit  `mapstructure:"buy"`
	SellLimit Limit  `mapstructure:"sell"`
}

type Limit struct {
	Amount     float64 `mapstructure:"amount"`
	Percentage float64 `mapstructure:"percentage"`
}

func main() {
	viper.SetConfigName("configuration")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Configuration
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}
	client := spot.NewREST()
	client.BaseURL = os.Getenv("KRAKEN_API_SPOT_REST_URL")
	client.PublicKey = os.Getenv("KRAKEN_API_SPOT_PUBLIC")
	client.PrivateKey = os.Getenv("KRAKEN_API_SPOT_SECRET")

	for _, pos := range config.Positions {
		ticker, err := client.Ticker(&spot.TickerRequest{
			Pair: pos.Name,
		})
		if err != nil {
			log.Print(err)
		}
		currentPrice := ticker.Result[pos.Name].Ask[0].Float64()
		
		buyLimit := currentPrice - (currentPrice * (pos.BuyLimit.Percentage/100))
		sellLimit := currentPrice + (currentPrice * (pos.SellLimit.Percentage/100))

		buyVol := fmt.Sprintf("%f", pos.BuyLimit.Amount/buyLimit)

		buyOrder, err := client.AddOrder(
			&spot.AddOrderRequest{
			Type: "buy",
			OrderType: "limit",
			Pair: pos.Name,
			Volume: buyVol,
			Price: fmt.Sprintf("%.2f", buyLimit),
		})
		if err != nil {
			log.Printf("%+v\n", buyOrder)
			log.Print(err)
		}

		if pos.SellLimit.Amount == 0 {
			pos.SellLimit.Amount = (sellLimit/buyLimit) * pos.BuyLimit.Amount
		}

		sellVol := fmt.Sprintf("%f", pos.SellLimit.Amount/sellLimit)

		sellOrder, err := client.AddOrder(
			&spot.AddOrderRequest{
			Type: "sell",
			OrderType: "limit",
			Pair: pos.Name,
			Volume: sellVol,
			Price: fmt.Sprintf("%.2f", sellLimit),
		})
		if err != nil {
			log.Printf("%+v\n", sellOrder)
			log.Print(err)
		}
	}
}
