package main

import (
	"beaconchain/rpc"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New().WithField("module", "cacher")

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
	logger.Printf("switched to %v\n", s.hosts[s.index])
	return s.Get()
}

func NewClients(hosts []string) (*clients, error) {
	s := &clients{
		index: 0,
		list:  make([]*rpc.PrysmClient, 0),
		hosts: hosts,
	}
	for _, host := range hosts {
		logger.Printf("connecting to RPC of %v", host)
		client, err := rpc.NewPrysmClient(host)
		if err != nil {
			logger.Printf("error connecting to %v: %v", host, err)
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
var offset = flag.Int("offset", 0, "in case of head, offset from the head")
var limit = flag.Int("limit", 1000, "max number of epochs to look at")
var timeout = flag.Int("timeout", 0, "time period, in minutes (watching head only). If takes less, process will wait for this period after job is done")
var debug = flag.Bool("debug", false, "do some debugging instead of the job")
var inc = flag.Bool("inc", false, "do through epochs incrementally")

var cacheBalances = flag.Bool("balances", true, "cache balances")
var cacheValidators = flag.Bool("validators", true, "cache validator lists")
var cacheAssignments = flag.Bool("assignments", true, "cache assignmenets")

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}
	flag.Parse()

	if *gethead {
		client, err := rpc.NewPrysmClient(*hosts)
		if err != nil {
			logger.Fatal(err)
		}
		head, err := client.GetChainHead()
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Print(head.HeadEpoch)
		os.Exit(0)
	}
	clients, err := NewClients(strings.Split(*hosts, ","))
	if err != nil {
		logger.Fatal(err)
	}

	if *debug {
		// epoch := uint64(49050)
		// pb, err := rpc.LoadAssignmentsPB(epoch)
		// if err != nil {
		// 	logger.Fatal(err)
		// }
		// out := rpc.NewAssignmentsFromPB(epoch, pb)
		// rpc.SaveAssignments(epoch, out)

		// pbjson, _ := json.MarshalIndent(pb, "", "  ")
		// outjson, _ := json.MarshalIndent(out, "", "  ")
		// if string(outjson) != string(outjson1) {
		// 	log.Fatal("codec strings mismatch")
		// }
		// // log.Println("codec strings full match")
		// fmt.Println(string(outjson))
		os.Exit(0)
	}

	estHeadEpoch := 0
	headEpoch := *head
	if headEpoch == 0 {
		head, err := clients.Get().GetChainHead()
		if err != nil {
			logger.Fatal(err)
		}
		headEpoch = int(head.HeadEpoch)
		estHeadEpoch = int(head.HeadEpoch)
		logger.Println("chain head epoch", head.HeadEpoch)
	}

	sign := -1
	if *inc {
		sign = 1
		if estHeadEpoch == 0 {
			head, err := clients.Get().GetChainHead()
			if err != nil {
				logger.Fatal(err)
			}
			estHeadEpoch = int(head.HeadEpoch)
		}
	}
	i := *offset
	failures := map[uint64]int{}
	since := time.Now()
	for {
		start := time.Now()
		cont := false
		epoch := uint64(int(headEpoch) + i*sign)

		if *cacheBalances {
			_, err := clients.Get().GetBalancesForEpoch(int64(epoch))
			if err != nil {
				if _, ok := failures[epoch]; !ok {
					failures[epoch] = 0
				}
				failures[epoch] += 1
				logger.Printf("[balances] epoch %d error: %v, took %v\n", epoch, err, time.Since(start))
				if failures[epoch] < clients.Len() {
					// try again on other server, otherwise skip
					clients.Next()
				} else {
					i++            // all hosts were requested, just skip to the next epoch
					clients.Next() // switch anyway to a better server
				}
				cont = true
			}
		}
		if *cacheValidators {
			_, err := clients.Get().GetEpochValidators(uint64(epoch))
			if err != nil {
				if _, ok := failures[epoch]; !ok {
					failures[epoch] = 0
				}
				failures[epoch] += 1
				logger.Printf("[validators] epoch %d error: %v, took %v\n", epoch, err, time.Since(start))
				if failures[epoch] < clients.Len() {
					// try again on other server, otherwise skip
					clients.Next()
				} else {
					i++            // all hosts were requested, just skip to the next epoch
					clients.Next() // switch anyway to a better server
				}
				cont = true
			}
		}

		if *cacheAssignments {
			_, err := clients.Get().GetEpochAssignments(epoch)
			if err != nil {
				if _, ok := failures[epoch]; !ok {
					failures[epoch] = 0
				}
				failures[epoch] += 1
				logger.Printf("[assignment] epoch %d error: %v, took %v\n", epoch, err, time.Since(start))
				if failures[epoch] < clients.Len() {
					// try again on other server, otherwise skip
					clients.Next()
				} else {
					i++            // all hosts were requested, just skip to the next epoch
					clients.Next() // switch anyway to a better server
				}
				cont = true
			}
		}
		if cont {
			continue
		}
		logger.Printf("epoch %d took %v", epoch, time.Since(start))

		i++
		nextEpoch := int(headEpoch) + sign*i
		if sign > 0 {
			logger.Printf("epoch %d took %v, est head %v", epoch, time.Since(start), estHeadEpoch)
			if i >= *limit || nextEpoch > estHeadEpoch {
				break
			}
		} else {
			if i >= *limit || nextEpoch == 0 {
				break
			}
		}

		if timeout != nil && *timeout > 0 {
			dur := time.Now().Sub(since)
			maxDuration := time.Duration(*timeout) * time.Minute
			if dur >= maxDuration {
				logger.Printf("this takes more than %v, EXITING", maxDuration)
				os.Exit(0)
			}
		}
	}

	if timeout != nil && *timeout > 0 {
		dur := time.Now().Sub(since)
		maxDuration := time.Duration(*timeout) * time.Minute
		if dur < maxDuration {
			logger.Printf("sleeping for %v", maxDuration-dur)
			time.Sleep(maxDuration - dur)
		}
	}
}
