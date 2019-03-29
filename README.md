# Visualize Blockchain with TigerGraph

The goal of this project was to provide a method for any user to upload blockchain data onto the TigerGraph database w/o the use of extra hardware boosts or tremendous amount of CPU usage or memory consumption. The reason for this was to make analyzing the blockchain available to both interested companies as well as amateur developers. Running this program will produce parsed sets of blockchain data that will be ready to analyze on the TigerGraph platform.

We provide a way to utilize TigerGraph's batch loading capabilites.

We provide an implementation of this code in Go.

## Getting started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* Go1.1+ installed
* [TigerGraph Developer or Enterprise Edition](https://www.tigergraph.com/download/)
* Any operating system should be fine, however the commands here are for OS X
* At least 5GB of RAM and 400GB of Disk Space (if creating csv files), 200GB of Disk Space (if using kafka)
* Apache Kafka (if planning to use Kafka loader)

#### Note:
TigerGraph default query timeout is 16. Please check [TigerGraph Documentation: Why has my query timed out?](https://docs.tigergraph.com/admin/admin-guide/system-administration-faqs#TigerGraphSystemAdministrationFAQs-Whyhasmyrequesttimedout?) to modify your timeout.

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
./blockchain_parse_csv \
  -batch=<size of batch> \
  -db=<database directory> \
  -end=<size of end block> \
  -output=<csv files output directory> \
  -start=<size of desired start block>
```
You can also use `./blockchain_parse_csv -help=true` or `./blockchain_parse_csv -h` for help.


## Loading and Visualizing the data into TigerGraph

You can find the GSQL queries under `/GSQL` directory. 

We can execute this full set of commands without entering the GSQL shell. For example, if the GSQL commands is in a Linux file named /home/tigergraph/hello2.gsql. In a Linux shell, under /home/tigergraph, type the following:

```
gsql hello2.gsql
```

1. Creating graph

Run following command in Linux shell.

```
gsql schema.gsql
```

Or directly copy the content in GSQL/schema.gsql into GSQL shell.

2. Creating loading job

If you haven't set up `sys.data_root` before, you should do step 3 before step 2. Then run following command in Linux shell.

```
gsql loading_job.gsql
```

Or directly copy the content in GSQL/loading_job.gsql into GSQL shell.


3. Setting the session parameter

Type gsql as below. A GSQL shell prompt should appear as below.

```
$ gsql 
GSQL >
```
Then set `sys.data_root`

```
SET sys.data_root=<absolute path of the csv files>
```

For example, if the csv files you get before are under `/home/tigergraph/bitcion_data`, you can input `SET sys.data_root="/home/tigergraph/bitcion_data"`

4. Running loading job 

```
RUN JOB load_blockchain_data
```
More detailed reference please check [TigerGraph document](https://docs.tigergraph.com/dev/gsql-ref/ddl-and-loading/creating-a-loading-job)

5. Visualization

Run following command in Linux shell.

```
gsql queries.gql
```

## Dataset
In our test, we used the full-node data by 2019.02.16, and after transforming to CSV files, the size is shown as below
```
$ du -h /mnt/data/csv/btc_csv/
569G    /mnt/data/csv/btc_csv/

$ du -h /mnt/data/csv/btc_csv/*
133M    /mnt/data/csv/btc_csv/blocks.csv
346G    /mnt/data/csv/btc_csv/input.csv
171G    /mnt/data/csv/btc_csv/output.csv
52G     /mnt/data/csv/btc_csv/transactions.csv
```
## TigerGraph Data Loading Report
### Hardware
In the test, we used AWS EC2 instance of type [r5.24xlarge](https://aws.amazon.com/ec2/instance-types/r5/), which has 96vCPU and 768G memory.

TigerGraph system and data storage are running on AWS EBS with 6144 IOPS

In the below section, the memory usage is only 200G. Theoratically, you can load the data with [TigerGraph developer edition](https://www.tigergraph.com/developer/) on any machine with 250G memory. However, some large queries would need more memory.

### Loading Time
The total loading time is 2 hours 50 minutes
The loading time of each file is shown as below:
```
  +----------------------------------------------------------------------------------+
  |                              FILENAME |   LOADED LINES |   AVG SPEED |   DURATION|
  |      /mnt/data/csv/btc_csv/blocks.csv |         563316 |    293 kl/s |     1.92 s|
  |       /mnt/data/csv/btc_csv/input.csv |      968416307 |    461 kl/s |  2096.50 s|
  |      /mnt/data/csv/btc_csv/output.csv |     1033945270 |    144 kl/s |  7149.25 s|
  |/mnt/data/csv/btc_csv/transactions.csv |      383614417 |    384 kl/s |   997.75 s|
  +----------------------------------------------------------------------------------+
```

### Memory Usage
```
$ free -h
              total        used        free      shared  buff/cache   available
Mem:           747G        202G         13G        872K        532G        540G
Swap:            0B          0B          0B
```

### Disk Usage
```
$ du /mnt/data/tigergraph/gstore/ -h
...
262G    /mnt/data/tigergraph/gstore/
```

## Contributors
```
Sai Pusapaty
Aining Wang
Heqing Ya (heqing.ya [at] tigergraph.com)
```
