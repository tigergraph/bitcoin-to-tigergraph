package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"encoding/hex"
	"github.com/piotrnar/gocoin/lib/btc"
	"github.com/piotrnar/gocoin/lib/others/blockdb"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Building structs to hold data
type BlockDat struct {
	dat []byte
	num int
	er  error
}

type TxnDat struct {
	block *btc.Block
	tx    *btc.Tx
}

func main() {

	// PARAMETERS
	batch_size := 1000
	end_block := 350000
	desired_start_block := 0//224000
	database := "/Users/sai/Desktop/bitcoin_data/blocks"

	log.Println("Starting the parsing process...")
	// Set real Bitcoin network
	Magic := [4]byte{0xF9, 0xBE, 0xB4, 0xD9}

	// Specify blocks directory
	BlockDatabase := blockdb.NewBlockDB(database, Magic)

	// Connecting to broker and creating new kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "192.168.0.111:9092"})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// make channels to pass data
	block_channel := make(chan BlockDat, 50000)
	txn_channel := make(chan TxnDat, 300000)

	// Create wait group so that goroutines can all finish before proceeding to next batch
	var tx_wg sync.WaitGroup

	// Process block data in channel in the background
	go func() {
		for msg := range block_channel {
			processBlock(msg, txn_channel, p, tx_wg)
		}
	}()

	// Process transaction data in channel in the background
	go func() {
		for msg := range txn_channel {
			processTxn(msg, p)
		}
	}()

	// loop through all the blocks
	for start_block := 0; start_block <= end_block; start_block++ {

		dat, er := BlockDatabase.FetchNextBlock()

		if start_block < desired_start_block {
			continue
		}

		// Store block data in channel
		block := BlockDat{dat: dat, er: er, num: start_block}
		block_channel <- block

		// Once batch size is reached process the rest of the txn channel and block channel before proceeding
		if start_block%batch_size == 0 {
			log.Println("Reached batch:", start_block/batch_size)
			for len(block_channel) > 0 {
				time.Sleep(10 * time.Millisecond)
			}

			log.Println("Cleaning up transactions...")
			for len(txn_channel) > 0 {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}

	tx_wg.Wait()

	log.Println("Closing Channels...")
	close(block_channel)
	close(txn_channel)

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)

	log.Println("FINISHED!")
}


func publish_message(p *kafka.Producer, msg string, topic string) {
	p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(msg),
	}, nil)
}

func processTxn(data TxnDat, p *kafka.Producer) {

	tx := data.tx
	bl := data.block

	txn_msg := fmt.Sprintf("%v,%v,%v,%v,%v,%v",
		tx.Hash.String(),
		tx.Size,
		tx.Lock_time,
		tx.Version,
		bl.Hash,
		tx.IsCoinBase(),
	)

	publish_message(p, txn_msg, "txn_data")

	if !tx.IsCoinBase() {
		for txin_index, txin := range tx.TxIn {
			prev_outid := strings.Split(txin.Input.String(), "-")
			outid, _ := strconv.Atoi(prev_outid[1])
			txn_outid := prev_outid[0] + "_" + strconv.Itoa(outid)
			input_msg := fmt.Sprintf("%v,%v,%v,%v,%v",
				txn_outid,
				txin_index,
				tx.Hash.String(),
				hex.EncodeToString(txin.ScriptSig),
				txin.Sequence,
			)
			publish_message(p, input_msg, "input_data")
		}
	}

	for txo_index, txout := range tx.TxOut {
		curr_outid := tx.Hash.String() + "_" + strconv.Itoa(txo_index)
		txout_addr := btc.NewAddrFromPkScript(txout.Pk_script, false)
		output_msg := fmt.Sprintf("%v,%v,%v,%v,%v",
			curr_outid,
			txo_index,
			txout.Value,
			txout_addr,
			tx.Hash.String(),
		)
		publish_message(p, output_msg, "output_data")
	}

}

func processBlock(data BlockDat, txn_channel chan TxnDat, p *kafka.Producer, tx_wg sync.WaitGroup) {

	i := data.num
	dat := data.dat
	er := data.er

	if dat == nil || er != nil {
		log.Println("END of DB file")
		return
	}

	bl, er := btc.NewBlock(dat[:])

	if er != nil {
		println("Block inconsistent:", er.Error())
		return
	}

	block_msg := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v",
		i,
		bl.Hash,
		hex.EncodeToString(bl.MerkleRoot()),
		reverse_flip_pairs(hex.EncodeToString(bl.ParentHash())),
		convert_to_time(bl.BlockTime()),
		bl.Version(),
		bl.Bits(),
	)

	publish_message(p, block_msg, "block_data")

	bl.BuildTxList()
	tx_wg.Add(len(bl.Txs))


	for _, tx := range bl.Txs {
		go func(tx *btc.Tx, bl *btc.Block) {

			defer tx_wg.Done()

			transaction := TxnDat{block: bl, tx: tx}
			txn_channel <- transaction
		}(tx, bl)
	}
}

func reverse_flip_pairs(parent_hash string) string {

	rs := []rune(parent_hash)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	for i := 1; i < len(rs); i += 2 {
		rs[i], rs[i-1] = rs[i-1], rs[i]
	}
	return string(rs)
}

func convert_to_time(sec_to_add uint32) time.Time {

	start := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	secs := strconv.Itoa(int(sec_to_add)) + "s"
	dur, _ := time.ParseDuration(secs)

	return start.Add(dur)
}