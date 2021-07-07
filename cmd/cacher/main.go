package main

import (
	"beaconchain/rpc"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type clients struct {
	index int
	list  []*rpc.PrysmClient
	hosts []string
}

func (s *clients) Get() *rpc.PrysmClient {
	return s.list[s.index]
}

func (s *clients) Len() int {
	return len(s.hosts)
}

func (s *clients) Next() *rpc.PrysmClient {
	s.index = (s.index + 1) % len(s.list)
	log.Printf("switched to %v\n", s.hosts[s.index])
	return s.Get()
}

func NewClients(hosts []string) (*clients, error) {
	s := &clients{
		index: 0,
		list:  make([]*rpc.PrysmClient, 0),
		hosts: hosts,
	}
	for _, host := range hosts {
		log.Printf("connecting to RPC of %v\n", host)
		client, err := rpc.NewPrysmClient(host)
		if err != nil {
			log.Printf("error connecting to %v: %v\n", host, err)
		} else {
			s.list = append(s.list, client)
		}
	}
	if len(s.list) == 0 {
		return nil, errors.New("no Prysm clients to connect")
	}
	return s, nil
}

var hosts = flag.String("hosts", "localhost:4000", "comma-separated list of hosts to connect to")
var gethead = flag.Bool("get-head", false, "return head of")
var head = flag.Int("head", 0, "block to start reading")
var limit = flag.Int("limit", 1000, "max number of epochs to look at")
var debug = flag.Bool("debug", false, "do some debugging instead of the job")

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.Parse()

	if *gethead {
		client, err := rpc.NewPrysmClient(*hosts)
		if err != nil {
			log.Fatal(err)
		}
		head, err := client.GetChainHead()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(head.HeadEpoch)
		os.Exit(0)
	}
	clients, err := NewClients(strings.Split(*hosts, ","))
	if err != nil {
		log.Fatal(err)
	}

	if *debug {
		epoch := uint64(801)
		out, err := clients.Get().GetEpochAssignments(epoch)
		if err != nil {
			log.Fatal(err)
		}
		outjson, _ := json.MarshalIndent(out, "", "  ")
		// s := rpc.NewAssignments(epoch, out)
		// b, _ := json.MarshalIndent(s.Get(), "", "  ")

		// if string(outjson) != string(b) {
		// 	log.Fatal("codec strings mismatch")
		// }
		// log.Println("codec strings full match")
		rpc.SaveAssignmentsMaps(epoch, out)
		out2, err := rpc.LoadAssignmentsMaps(epoch)
		if err != nil {
			log.Fatal(err)
		}
		b, _ := json.MarshalIndent(out2, "", "  ")
		if string(outjson) != string(b) {
			log.Fatal("codec strings mismatch")
		}
		log.Println("codec strings full match")
		fmt.Println(string(b))
		os.Exit(0)
	}

	headEpoch := *head
	if headEpoch == 0 {
		head, err := clients.Get().GetChainHead()
		if err != nil {
			log.Fatal(err)
		}
		headEpoch = int(head.HeadEpoch)
		fmt.Println("chain head epoch", head.HeadEpoch)
	}

	i := 0
	failures := map[uint64]int{}
	for {
		start := time.Now()

		epoch := uint64(int(headEpoch) - i)
		// _, err := clients.Get().GetBalancesForEpoch(epoch)
		_, err := clients.Get().GetEpochAssignments(epoch)
		if err != nil {
			if _, ok := failures[epoch]; !ok {
				failures[epoch] = 0
			}
			failures[epoch] += 1
			log.Printf("epoch %d error: %v, took %v\n", epoch, err, time.Since(start))
			if failures[epoch] < clients.Len() {
				// try again on other server, otherwise skip
				clients.Next()
			} else {
				i++            // all hosts were requested, just skip to the next epoch
				clients.Next() // switch anyway to a better server
			}
			continue
		}
		log.Printf("epoch %d took %v\n", epoch, time.Since(start))

		i++
		if i >= *limit || (int(headEpoch)-i) == 0 {
			break
		}
	}
}
