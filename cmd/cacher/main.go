package main

import (
	"beaconchain/rpc"
	"beaconchain/types"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func workerA(client *rpc.PrysmClient, epoch uint64) {
	start := time.Now()
	key := fmt.Sprintf("res/e%08x.a", epoch)
	if client.Storage.Has(key) {
		fmt.Printf("[skip] worker A%v, took %s\n", epoch, time.Since(start))
		return
	}
	res, err := client.GetEpochAssignments(epoch)
	if err != nil {
		log.Println("[a] error: ", err)
	}
	out, _ := json.Marshal(res)
	if err := client.Storage.Set(key, out); err != nil {
		log.Println("[warn]", err)
	}
	fmt.Printf("[done] worker A%v in %s\n", epoch, time.Since(start))
}

func workerB(client *rpc.PrysmClient, epoch uint64) {
	start := time.Now()
	key := fmt.Sprintf("res/e%08x.b", epoch)
	if client.Storage.Has(key) {
		fmt.Printf("[skip] worker B%v, took %s\n", epoch, time.Since(start))
		return
	}
	res, err := client.GetBalancesForEpoch(int64(epoch))
	if err != nil {
		log.Println("[b] error: ", err)
	}
	out, _ := json.Marshal(res)
	if err := client.Storage.Set(key, out); err != nil {
		log.Println("[warn]", err)
	}
	fmt.Printf("[done] worker B%v, took %s\n", epoch, time.Since(start))
}

func workerS(client *rpc.PrysmClient, epoch uint64) {
	start := time.Now()
	key := fmt.Sprintf("res/e%08x.s", epoch)
	if client.Storage.Has(key) {
		fmt.Printf("[skip] worker S%v, took %s\n", epoch, time.Since(start))
		return
	}

	slots := map[uint64][]*types.Block{}
	for slot := epoch*32 + 1; slot <= (epoch+1)*32; slot++ {
		val, err := client.GetBlocksBySlot(slot)
		if err != nil {
			log.Println("[a] error: ", err, "slot", slot)
			continue
		}
		slots[slot] = val
	}
	out, _ := json.Marshal(slots)
	if err := client.Storage.Set(key, out); err != nil {
		log.Println("[warn]", err)
	}
	fmt.Printf("[done] worker S%v, took %s\n", epoch, time.Since(start))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cacher, err := rpc.NewStorage()
	if err != nil {
		log.Fatal(err)
	}
	client, err := rpc.NewPrysmClient("localhost:4000", cacher)
	if err != nil {
		log.Fatal(err)
	}
	head, err := client.GetChainHead()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chain head epoch", head.HeadEpoch)

	// wg := sync.WaitGroup{}
	chA := make(chan int, 2)
	chB := make(chan int, 2)
	chS := make(chan int, 2)
	go func() {
		for eA := range chA {
			workerA(client, uint64(eA))
		}
	}()
	go func() {
		for eB := range chB {
			workerB(client, uint64(eB))
		}
	}()
	go func() {
		for eS := range chS {
			workerS(client, uint64(eS))
		}
	}()

	for epoch := 1; epoch < int(head.HeadEpoch); epoch++ {
		// fmt.Println("[queued] epoch", epoch)
		chA <- epoch
		chB <- epoch
		chS <- epoch
	}
	close(chA)
}
