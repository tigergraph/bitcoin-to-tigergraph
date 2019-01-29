# Visualize Blockchain with TigerGraph

The goal of this project was to provide a method for any user to upload blockchain data onto the TigerGraph database w/o the use of extra hardware boosts or tremendous amount of CPU usage or memory consumption. The reason for this was to make analyzing the blockchain available to both interested companies as well as amateur developers. Running this program will produce parsed sets of blockchain data that will be ready to analyze on the TigerGraph platform.

We provide a way to utilize TigerGraph's batch loading capabilites as well as it's recently developed streaming capabilities via kafka loaders.

We provide an implementation of this code in both Python and Go.

## Getting started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* Python3 or Go1.1+ installed
* [TigerGraph Developer or Enterprise Edition](https://www.tigergraph.com/download/)
* Any operating system should be fine, however the commands here are for OS X
* At least 5GB of RAM and 400GB of Disk Space (if creating csv files), 200GB of Disk Space (if using kafka)
* Apache Kafka (if planning to use Kafka loader)

### Getting the raw blockchain data

Fortunately [Bitcoin Core](https://bitcoin.org/en/download) is public software that downloads all of blockchain data directly. Please note that downloading all of the files is a lengthy process and will generate ~200GB of data (as of January 2019). The data will be in the *.dat format within the "blocks" folder.

## Golang setup

## Installing dependencies

Used this [gocoin library](github.com/piotrnar/gocoin) to parse the blockchain. 

```
go get -u github.com/piotrnar/gocoin
```

Please build all the files that are necessary by cd'ing into appropriate folders and using 

`go build`

In the program, set the desired start block and desired end block and press run. Other factors to change if wanted are batch size (determines number of goroutines) as well as buffer sizes for each channel used to process data. You can expect this program to take a few hours to process all of blockhain with given parameters.

## Running

Please run blockchain_parse.go


## Python setup (NOT FINISHED)

### Installing dependencies

The [pyblockchain library](https://github.com/toidi/pyblockchain) was used to read the .dat files. It can be installed with

```
pip3 install pyblockchain
``` 

Unfortunately as the pyblockchain library contains no way to count the overall block number, there is a need to calculate it separately. A csv file is already provided up till file number 975 with block counts. If this number needs to be updated, please run **count_blocks.py** (this implementation uses kafka for faster counting, but is not necessary).

### Running to create csv files

Run parse_blockchain.py

### Runnning to kakfka

Run parse_blockchain_kafka.py


## Putting the data into TigerGraph

### Create graph

In GSQL please run the commands found in create_graph.txt

### Using .csv files

Data can be loaded through two different ways. One can be by creating a loading job using the commands in "load_data.txt" in GSQL shell. 
