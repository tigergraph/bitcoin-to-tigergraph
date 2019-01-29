from blockchain.reader import BlockchainFileReader
import multiprocessing as mp
from kafka import KafkaConsumer, KafkaProducer
import csv


seen = set()
with open("file_block_count.csv","r") as f:
	reader = csv.reader(f, delimiter=",")
	for line in reader:
		seen.add(int(line[0]))

print(seen, len(seen))



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
        _producer = KafkaProducer(bootstrap_servers=['localhost:9092'], api_version=(0, 10))
    except Exception as ex:
        print('Exception while connecting Kafka')
        print(str(ex))
    finally:
        return _producer


def worker(file_no):

	kafka_producer = connect_kafka_producer()

	block_no = None

	if file_no % 50 == 0:
		print(file_no)

	str_file_no = str(file_no)
	while len(str_file_no) < 5:
		str_file_no = "0" + str_file_no

	try:
		block_reader = BlockchainFileReader('/Users/sai/Desktop/bitcoin_data/blocks/blk'+str_file_no+'.dat')

		for no,block in enumerate(block_reader):

			block_no = no

		val = [file_no,block_no]
		publish_message(kafka_producer,"file_block_count",','.join(map(str,val)))
	except Exception as e:
		print(file_no,block_no,"caused error")
		print(e)

def handler(start,num_of_files):
	data = [x for x in range(start,num_of_files) if x not in seen]
	p = mp.Pool(8)
	p.map(worker, data)

if __name__ == "__main__":

	handler(975,977)

