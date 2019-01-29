from kafka import KafkaConsumer
import csv

consumer = KafkaConsumer("file_block_count", auto_offset_reset='earliest',
                             bootstrap_servers=['localhost:9092'], api_version=(0, 10), consumer_timeout_ms=1000)

seen = set()
with open("/Users/sai/Desktop/file_block_count.csv","r") as f:
	reader = csv.reader(f, delimiter=",")
	for line in reader:
		seen.add(int(line[0]))

file_start = 0
file_stop = 1360

while len(seen) < (file_stop-file_start):

	if len(seen) % 100 == 0:
		print(len(seen), "records have been processed")

	for msg in consumer:

		file_block_count_csv = open('/Users/sai/Desktop/file_block_count.csv', 'a')
		fbc_fw = csv.writer(file_block_count_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)

		val = msg.value.decode(encoding='utf-8').split(",")
		if val[0] not in seen and int(val[0]) >= file_start:
			fbc_fw.writerow(val)
			seen.add(val[0])
print("DONE")



