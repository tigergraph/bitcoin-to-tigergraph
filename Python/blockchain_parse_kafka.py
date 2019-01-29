from blockchain.reader import BlockchainFileReader
import multiprocessing as mp
from kafka import KafkaConsumer, KafkaProducer
import csv
import time

def cumulative_counts(last_file_no):
	counts = {}
	origin = {}
	with open("file_block_count.csv","r") as f:
		reader = csv.reader(f, delimiter=",")
		for line in reader:
			counts[int(line[0])] = int(line[1])
			origin[int(line[0])] = int(line[1])

	print(origin[0])
	print(counts[0])

	for i in range(1,last_file_no+1):
		counts[i] += counts[i-1] + 1
	for j in range(last_file_no+1):
		counts[j] -= origin[j]

	return counts

def publish_message(producer_instance, topic_name, value):
    try:
        value_bytes = bytes(value, encoding='utf-8')
        producer_instance.send(topic_name, value=value_bytes)
        producer_instance.flush()
    except Exception as ex:
        print('Exception in publishing message')
        print(str(ex))


def connect_kafka_producer():
    _producer = None
    try:
        _producer = KafkaProducer(bootstrap_servers=['127.0.0.1:9092'], api_version=(0, 10))
    except Exception as ex:
        print('Exception while connecting Kafka')
        print(str(ex))
    finally:
        return _producer

def worker(file_no):

	t1 = time.time()

	kafka_producer = connect_kafka_producer()

	t2 = time.time()
	print("connecting to kafka producer:",t2-t1)
	try:
		num_string = str(file_no)
		while len(num_string) != 5:
			num_string = "0" + num_string

		block_reader = BlockchainFileReader('/Users/sai/Desktop/bitcoin_data/blocks/blk'+num_string+'.dat')

		t3 = time.time()
		print("reading .dat",t3-t2)

		for no,block in enumerate(block_reader):

			t4 = time.time()

			if no % 50 == 0:
				print("Reading block #",no,"in file #",file_no)

			current_block_hash = block.header.merkle_hash
			prev_block_hash = block.header.previous_hash

			block_size = block.header.block_size
			block_version = block.header.version
			nonce = block.header.nonce
			bits = block.header.bits
			block_timestamp = block.header.time
			block_id = no + counts[file_no]
			prev_block_id = block_id - 1

			val = [current_block_hash,prev_block_hash,block_size,block_version,nonce,bits,block_timestamp,block_id,prev_block_id]
			publish_message(kafka_producer,"blocks",','.join(map(str,val)))

			t5 = time.time()
			print("publishing",t5-t4)

			transactions = block.transactions

			for transaction in transactions:

				total_value = 0
				txn_hash = transaction.txn_hash
				txn_version = transaction.version
				txt_timestamp = transaction.lock_time

				outputs = transaction.outputs

				for i,output in enumerate(outputs):

					outid = str(i)
					txn_hash_outid = txn_hash+"_"+outid
					output_val = output.value
					try:
						address = output.address
					except:
						address = "None"
					total_value += output_val

					val = [txn_hash_outid,outid,output_val,address,txn_hash]
					publish_message(kafka_producer,"outputs",','.join(map(str,val)))

				inputs = transaction.inputs

				for i, inp in enumerate(inputs):

					prev_txn_hash = inp.previous_hash
					txn_hash_outid = prev_txn_hash+"_"+str(inp.txn_out_id)
					outid = inp.txn_out_id

					val = [txn_hash_outid,outid,txn_hash]
					publish_message(kafka_producer,"ingoing_payment",','.join(map(str,val)))


				val = [txn_hash,total_value,txn_version,txt_timestamp,current_block_hash]
				publish_message(kafka_producer,"transactions",','.join(map(str,val)))

			t6 = time.time()
			print("reading transactions",t6-t5)

		if kafka_producer is not None:
			kafka_producer.close()

		
		
		
		
	except:
		print("Error in streaming data")

def handler(last_file_no):
	data = [x for x in range(900,last_file_no+1)]
	p = mp.Pool(8)
	p.map(worker, data)

if __name__ == '__main__':

	last_file_no = 975
	counts = cumulative_counts(last_file_no)
	handler(last_file_no)





