package main

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/piotrnar/gocoin/lib/btc"
	"github.com/piotrnar/gocoin/lib/others/blockdb"
	"log"
	"os"
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
	end_block := 50000
	desired_start_block := 0
	database := "/Users/sai/Desktop/bitcoin_data/blocks"

	log.Println("Starting the parsing process...")
	// Set real Bitcoin network
	Magic := [4]byte{0xF9, 0xBE, 0xB4, 0xD9}

	// Specify blocks directory
	BlockDatabase := blockdb.NewBlockDB(database, Magic)

	//.Preparing files to write to
	blk_fl, err1 := os.OpenFile("blocks.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	txn_fl, err2 := os.OpenFile("transactions.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	in_fl, err3 := os.OpenFile("input.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	out_fl, err4 := os.OpenFile("output.csv", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)

	for _, err := range []error{err1, err2, err3, err4} {
		checkError("Cannot create file", err)
	}

	defer blk_fl.Close()
	defer txn_fl.Close()
	defer in_fl.Close()
	defer out_fl.Close()

	block_writer := csv.NewWriter(blk_fl)
	txn_writer := csv.NewWriter(txn_fl)
	in_writer := csv.NewWriter(in_fl)
	out_writer := csv.NewWriter(out_fl)

	defer block_writer.Flush()
	defer txn_writer.Flush()
	defer in_writer.Flush()
	defer out_writer.Flush()

	// make channels to pass data
	block_channel := make(chan BlockDat, 50000)
	txn_channel := make(chan TxnDat, 300000)

	// Create wait group so that goroutines can all finish before proceeding to next batch
	var tx_wg sync.WaitGroup

	// Process block data in channel in the background
	go func() {
		for msg := range block_channel {
			processBlock(msg, txn_channel, block_writer, tx_wg)
		}
	}()

	// Process transaction data in channel in the background
	go func() {
		for msg := range txn_channel {
			processTxn(msg, txn_writer, in_writer, out_writer)
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

	log.Println("FINISHED!")
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func processTxn(data TxnDat, txn_writer *csv.Writer, in_writer *csv.Writer, out_writer *csv.Writer) {

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

	txn_writer.Write(strings.Split(txn_msg, ","))

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
			in_writer.Write(strings.Split(input_msg, ","))
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
		out_writer.Write(strings.Split(output_msg, ","))
	}

}

func processBlock(data BlockDat, txn_channel chan TxnDat, block_writer *csv.Writer, tx_wg sync.WaitGroup) {

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

	block_writer.Write(strings.Split(block_msg, ","))

	bl.BuildTxList()
	tx_wg.Add(len(bl.Txs))

	//fmt.Println("Number of transactions for this block ",len(bl.Txs))

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
