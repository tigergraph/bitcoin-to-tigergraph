# Visualize Blockchain with TigerGraph

The goal of this project was to provide a method for any user to upload blockchain data onto the TigerGraph database w/o the use of extra hardware boosts or tremendous amount of CPU usage or memory consumption. The reason for this was to make analyzing the blockchain available to both interested companies as well as amateur developers. Running this program will produce parsed sets of blockchain data that will be ready to analyze on the TigerGraph platform.

We provide a way to utilize TigerGraph's batch loading capabilites as well as it's recently developed streaming capabilities via kafka loaders.

We provide an implementation of this code in both Python and Go.

## Getting started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* Go1.1+ installed
* [TigerGraph Developer or Enterprise Edition](https://www.tigergraph.com/download/)
* Any operating system should be fine, however the commands here are for OS X
* At least 5GB of RAM and 400GB of Disk Space (if creating csv files), 200GB of Disk Space (if using kafka)
* Apache Kafka (if planning to use Kafka loader)

### Getting the raw blockchain data

Fortunately [Bitcoin Core](https://bitcoin.org/en/download) is public software that downloads all of blockchain data directly. Please note that downloading all of the files is a lengthy process and will generate ~200GB of data (as of January 2019). The data will be in the *.dat format within the "blocks" folder.

Approximately it takes 3 days on with AWS t2.2xlarge instance. Block files are ~200GB and index is around 250G.

## Turning raw data into csv files

### Installing Golang dependencies

Used this [gocoin library](github.com/piotrnar/gocoin) to parse the blockchain. 

```
go get -u github.com/piotrnar/gocoin
```

Please build all the files that are necessary by cd'ing into appropriate folders and using `go build`

### Running

In the program, set the desired start block and desired end block and press run. Other factors to change if wanted are batch size (determines number of goroutines) as well as buffer sizes for each channel used to process data. You can expect this program to take a few hours to process all of blockhain with given parameters.

```
./blockchain_parse_csv -batch=<size of batch> -db=<database directory> -end=<size of end block> -output=<csv files output directory> -start=<size of desired start block>
```
You can also use `./blockchain_parse_csv -help=true` or `./blockchain_parse_csv -h` for help.



## Loading the data into TigerGraph

Please open GSQL and running following query.

1. Creating graph

```
create vertex Output (primary_id transaction_hash_outid string, outid int, transaction_value int)
create vertex Transaction (primary_id transaction_hash string, txn_size int, version_no int, txn_locktime int, is_coinbase bool)
create vertex Block (primary_id curr_hash string, block_index int, merkle_root string, block_time datetime, block_version int, block_bits int)
create vertex Address (primary_id address string)
create directed edge output_to_address (from Output, to Address) with reverse_edge="address_to_output"
create directed edge txn_output (from Transaction, to Output) with reverse_edge="output_origin_txn"
create directed edge txn_input (from Output, to Transaction) with reverse_edge="txn_origin_input"
create directed edge txn_to_block (from Transaction, to Block) with reverse_edge="block_to_txn"
create directed edge chain (from Block, to Block) with reverse_edge="reverse_chain"
create graph Block_Chain (*)
```

2. Creating loading job

```
begin
CREATE LOADING JOB load_blockchain_data FOR GRAPH Block_Chain {
	DEFINE FILENAME blocks = "$sys.data_root/blocks.csv";
	DEFINE FILENAME txns = "$sys.data_root/transactions.csv";
	DEFINE FILENAME outputs = "$sys.data_root/output.csv";
	DEFINE FILENAME ingoing_payment = "$sys.data_root/input.csv";
	LOAD blocks to vertex Block values ($1,$0,$2,$4,$5,$6);
	LOAD txns to vertex Transaction values ($0,$1,$3,$2,$5);
	LOAD blocks to edge chain values ($3,$1);
	LOAD txns to edge txn_to_block values ($0,$4);
	LOAD outputs to vertex Output values ($0,$1,$2);
	LOAD outputs to edge output_to_address values ($0,$3);
	LOAD outputs to edge txn_output values ($4,$0);
	LOAD ingoing_payment to edge txn_input values ($0,$2);
	LOAD outputs to vertex Address values ($3);
}
end
```

3. Setting the session parameter

```
SET sys.data_root = <absolute path of the csv files>
```
For example, if the csv files you get before are under `/home/tigergraph/bitcion_data`, you can input `SET sys.data_root = "/home/tigergraph/bitcion_data"`

4. Running loading job 

```
RUN JOB load_blockchain_data
```
More detailed reference please check [TigerGraph document](https://docs.tigergraph.com/dev/gsql-ref/ddl-and-loading/creating-a-loading-job)