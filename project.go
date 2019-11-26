package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sort"

	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index         int
	Timestamp     string
	BPM           string
	Hash          string
	PrevHash      string
	Miner         string
	blockdatahash string
	prevdatahash  string
	transaction   []string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block
var tempBlocks []Block
var oddBlocks = make(map[string]Block)
var evenBlocks = make(map[string]Block)
var towait int64

// candidateBlocks handles incoming blocks for validation
var candidateBlocks = make(chan Block)

// announcements broadcasts winning Miner to all nodes
var announcements = make(chan string)

var mutex = &sync.Mutex{}

// Miners keeps track of open Miners and balances
var Miners = make(map[string]int)

var turn int
var arrayaddress = make(map[string]int)
var oddaddress = make(map[string]int)
var oddtime = make(map[string]float64)
var eventime = make(map[string]float64)
var oddvalue []float64
var evenvalue []float64
var evenaddress = make(map[string]int)
var total int
var resultfile, _ = os.Create("result.txt")

// var written bool
var randomtime int
var mintime int

// var resultfile
// var index int

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// index = -1
	// resultfile, err := os.Create("result.txt")
	// check(err)

	total = 0
	turn = 0
	randomtime = 10
	towait = 10
	mintime = 0
	// written = false
	// create genesis block
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), "", calculateBlockHash(genesisBlock), "", "", "", "", []string{}}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	httpPort := os.Getenv("PORT")

	// start TCP and serve TCP server
	server, err := net.Listen("tcp", ":"+httpPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP Server Listening on port :", httpPort)
	defer server.Close()

	go func() {
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner()
			if (turn+1)%50 == 0 {
				mintime = randomtime
				randomtime = randomtime * 2
			}
		}
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

// pickWinner creates a lottery pool of Miners and chooses the Miner who gets to forge a block to the blockchain
// by random selecting from the pool, weighted by amount of tokens staked
func pickWinner() {
	start := time.Now()
	time.Sleep(time.Duration(30) * time.Second)
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()

	fmt.Println("Length of the tempBlocks")
	fmt.Println(len(temp))
	if len(temp) > 3 {
		// if the turn is of odd
		if (turn+1)%2 == 1 {
			for _, block := range temp {
				_, ok := oddaddress[block.Miner]
				if ok {
					oddBlocks[block.Miner] = block
					oddtime[block.Miner] = rand.Float64()*(float64(randomtime)-float64(mintime)) + float64(mintime+1)
					oddvalue = append(oddvalue, oddtime[block.Miner])
				}
			}
			fmt.Println("Chance of Odd Id Miners")
		} else {
			for _, block := range temp {
				_, ok := evenaddress[block.Miner]
				if ok {
					evenBlocks[block.Miner] = block
					eventime[block.Miner] = rand.Float64()*(float64(randomtime)-float64(mintime)) + float64(mintime+1)
					evenvalue = append(evenvalue, eventime[block.Miner])
				}
			}
			fmt.Println("Chance of even ID miners")
		}

		// time to pick odd wineer
		if (turn+1)%2 == 1 {
			sort.Float64s(oddvalue)
			towait = int64(oddvalue[0])
			time.Sleep(time.Duration(int64(oddvalue[0])) * time.Second)
			value := oddvalue[0]
			var key string
			var validate []string
			for k, v := range oddtime {
				if v == value {
					key = k
					// break
				} else {
					validate = append(validate, k)
				}
			}
			// Miner announcement
			for _ = range Miners {
				announcements <- "\n Winning Miner: " + oddBlocks[key].Miner + "\n"
			}

			fmt.Println("Winning miner :" + oddBlocks[key].Miner)
			fmt.Println("New Block verfication started ")
			// validating transaction
			var counttrue int
			var countfalse int
			counttrue = 0
			countfalse = 0
			for validatifyer := range validate {
				decision := transactionverify(oddBlocks[key], key, validate[validatifyer])
				if decision {
					counttrue++
				} else {
					countfalse++
				}
			}

			if counttrue > countfalse {
				mutex.Lock()
				Blockchain = append(Blockchain, oddBlocks[key])
				mutex.Unlock()
				fmt.Println("New Block is successfully added to the Blockchain \n")
			} else {
				fmt.Println("Mined Block is not verified")
			}
		} else {
			// pick the winner from even
			sort.Float64s(evenvalue)
			towait = int64(evenvalue[0])
			time.Sleep(time.Duration(int64(evenvalue[0])) * time.Second)

			value := evenvalue[0]
			var validate []string
			var key string
			for k, v := range eventime {
				if v == value {
					key = k
				} else {
					validate = append(validate, k)
				}
			}
			// Miner announcement
			for _ = range Miners {
				announcements <- "\n Winning Miner: " + evenBlocks[key].Miner + "\n"
			}

			fmt.Println("Winning miner :" + evenBlocks[key].Miner)
			fmt.Println("New Block verfication started ")
			// validating transaction
			var counttrue int
			var countfalse int
			counttrue = 0
			countfalse = 0
			for vaval := range validate {
				decision := transactionverify(evenBlocks[key], key, validate[vaval])
				if decision {
					counttrue++
				} else {
					countfalse++
				}
			}

			if counttrue > countfalse {
				mutex.Lock()
				Blockchain = append(Blockchain, evenBlocks[key])
				mutex.Unlock()
				// for _ = range Miners {
				fmt.Println("New Block is successfully added to the Blockchain \n")
				// }
			} else {
				// for _ = range Miners {
				// announcements <- "\n Verification result is false \n"
				fmt.Println("Mined Block is not verified")
				// }
			}
		}
		turn = turn + 1

		mutex.Lock()
		s := " Blocked mined: " + strconv.Itoa(len(Blockchain)-1) + " Duration: " + time.Since(start).String() + "\n"
		resultfile.WriteString(s)
		mutex.Unlock()

		mutex.Lock()
		oddvalue = []float64{}
		evenvalue = []float64{}
		for k := range oddtime {
			delete(oddtime, k)
		}
		for k := range eventime {
			delete(eventime, k)
		}
		for k := range oddBlocks {
			delete(oddBlocks, k)
		}
		for k := range evenBlocks {
			delete(evenBlocks, k)
		}
		tempBlocks = []Block{}
		mutex.Unlock()
	} else {
		fmt.Println("Number of miners must be greater than or equal to 4 then this will work")
		mutex.Lock()
		oddvalue = []float64{}
		evenvalue = []float64{}
		for k := range oddtime {
			delete(oddtime, k)
		}
		for k := range eventime {
			delete(eventime, k)
		}
		for k := range oddBlocks {
			delete(oddBlocks, k)
		}
		for k := range evenBlocks {
			delete(evenBlocks, k)
		}
		tempBlocks = []Block{}
		mutex.Unlock()
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()
	// Miner address
	var address string

	// allow user to allocate number of tokens to stake
	// the greater the number of tokens, the greater chance to forging a new block
	io.WriteString(conn, "Enter token balance:")
	scanBalance := bufio.NewScanner(conn)
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		t := time.Now()
		address = calculateHash(t.String())
		Miners[address] = balance
		if (total+1)%2 == 1 {
			oddaddress[address] = 1
			fmt.Println("Adding address in odd block")
		} else {
			evenaddress[address] = 1
			fmt.Println("Adding address in even block")
		}
		arrayaddress[address] = 1
		total = total + 1
		fmt.Println(Miners)
		break
	}

	for {
		mutex.Lock()
		oldLastIndex := Blockchain[len(Blockchain)-1]
		mutex.Unlock()

		mutex.Lock()
		dat, err := ioutil.ReadFile("data/text.txt")
		mutex.Unlock()
		bpm := string(dat)
		bpm = calculateHash(bpm)

		newBlock, err := generateBlock(oldLastIndex, bpm, address)
		if err != nil {
			log.Println(err)
		}
		if isBlockValid(newBlock, oldLastIndex) {
			candidateBlocks <- newBlock
		}

		// _, err = conn.(syscall.Conn).SyscallConn()

		// if err != nil {
		// 	fmt.Println("Connection is closed")
		// }

		time.Sleep(time.Duration(30+towait) * time.Second)

		mutex.Lock()
		output, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output)+"\n")
		io.WriteString(conn, "\n New Block mining round will start now \n")

		// _, err = conn.(syscall.Conn).SyscallConn()

		// if err != nil {
		// 	fmt.Println("Connection is closed")
		// }
		// written = false
	}
}

func transactionverify(newblock Block, miner string, validator string) bool {
	prevhash := newblock.prevdatahash
	for trans := range newblock.transaction {
		bpm := calculateHash(newblock.transaction[trans])
		prevhash = calculateHash(prevhash + bpm)
	}
	if prevhash == newblock.blockdatahash {
		return true
	} else {
		return false
	}
	return true
}

// isBlockValid makes sure block is valid by checking index
// and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	return calculateHash(record)
}

// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, BPM string, address string) (Block, error) {

	var newBlock Block
	t := time.Now()
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Miner = address
	newBlock.transaction = []string{}
	newBlock.prevdatahash = oldBlock.blockdatahash

	prevhash := oldBlock.blockdatahash
	for i := 1; i <= 50; i++ {
		mutex.Lock()
		if i < 10 {
			dat, err := ioutil.ReadFile("data/text0" + strconv.Itoa(i) + ".txt")
			mutex.Unlock()
			if err != nil {
				log.Println(err)
			}
			bpm := string(dat)
			newBlock.transaction = append(newBlock.transaction, bpm)
			bpm = calculateHash(bpm)
			prevhash = prevhash + bpm
			prevhash = calculateHash(prevhash)
		} else {
			dat, err := ioutil.ReadFile("data/text" + strconv.Itoa(i) + ".txt")
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()
			bpm := string(dat)
			newBlock.transaction = append(newBlock.transaction, bpm)
			bpm = calculateHash(bpm)
			prevhash = prevhash + bpm
			prevhash = calculateHash(prevhash)
		}
	}
	newBlock.blockdatahash = prevhash
	return newBlock, nil
}
