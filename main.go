package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/markcheno/go-quote"
	"github.com/markcheno/go-talib"
	"golang.org/x/sync/semaphore"
	"gopkg.in/ini.v1"
)

// Semaphore

var s *semaphore.Weighted = semaphore.NewWeighted(1)

func longProcess(ctx context.Context) {
	// isAcquire := s.TryAcquire(1)
	// if !isAcquire {
	//  fmt.Println("Could not get lock")
	//  return
	// }

	if err := s.Acquire(ctx, 1); err != nil {
		fmt.Println(err)
		return
	}
	defer s.Release(1)
	fmt.Println("Wait...")
	time.Sleep(1 * time.Second)
	fmt.Println("Done")
}

// iniでConfigの設定ファイルを読み込む

type ConfigList struct {
	Port      int
	DbName    string
	SQLDriver string
}

var Config ConfigList

func init() {
	cfg, _ := ini.Load("config.ini")
	Config = ConfigList{
		Port:      cfg.Section("web").Key("port").MustInt(),
		DbName:    cfg.Section("db").Key("name").MustString("example.sql"),
		SQLDriver: cfg.Section("db").Key("driver").String(),
	}
}

// JSON-RPC 2.0 over WebSocketでBitcoinの価格をリアルタイムに取得する

type JsonRPC2 struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Result  interface{} `json:"result,omitempty"`
	Id      *int        `json:"id,omitempty"`
}
type SubscribeParams struct {
	Channel string `json:"channel"`
}

func main() {
	// Semaphore

	ctx := context.TODO()
	go longProcess(ctx)
	go longProcess(ctx)
	go longProcess(ctx)
	time.Sleep(2 * time.Second)
	go longProcess(ctx)
	time.Sleep(5 * time.Second)

	// iniでConfigの設定ファイルを読み込む

	fmt.Printf("%T %v\n", Config.Port, Config.Port)
	fmt.Printf("%T %v\n", Config.DbName, Config.DbName)
	fmt.Printf("%T %v\n", Config.SQLDriver, Config.SQLDriver)

	// talibで株価分析する

	spy, _ := quote.NewQuoteFromYahoo(
		"spy", "2018-04-01", "2019-01-01", quote.Daily, true)
	fmt.Println(spy.CSV())
	rsi2 := talib.Rsi(spy.Close, 2)
	fmt.Println(rsi2)
	mva := talib.Ema(spy.Close, 14)
	fmt.Println(mva)

	// JSON-RPC 2.0 over WebSocketでBitcoinの価格をリアルタイムに取得する

	u := url.URL{Scheme: "wss", Host: "ws.lightstream.bitflyer.com", Path: "/json-rpc"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	if err := c.WriteJSON(&JsonRPC2{Version: "2.0", Method: "subscribe", Params: &SubscribeParams{"lightning_ticker_BTC_JPY"}}); err != nil {
		log.Fatal("subscribe:", err)
		return
	}

	for {
		message := new(JsonRPC2)
		if err := c.ReadJSON(message); err != nil {
			log.Println("read:", err)
			return
		}

		if message.Method == "channelMessage" {
			log.Println(message.Params)
		}
	}
}
