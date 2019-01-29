from blockchain.reader import BlockchainFileReader
import csv
import multiprocessing as mp


def worker(file_no, q1, q2, q3, q4, q5):

	str_file_no = str(file_no)
	while len(str_file_no) < 5:
		str_file_no = "0" + str_file_no

	block_reader = BlockchainFileReader('/Users/sai/Desktop/bitcoin_data/blocks/blk'+str_file_no+'.dat')
	old_curr_hash = None
	block_no = None
	c_blk_hash = None

	for no,block in enumerate(block_reader):

		if no % 1000 == 0:
			print("reading block #",no,"in file #",file_no)


		block_no = no
		current_block_hash = block.header.merkle_hash
		c_blk_hash = current_block_hash
		if not old_curr_hash:
			prev_block_hash = block.header.previous_hash
		else:
			prev_block_hash = old_curr_hash
		block_size = block.header.block_size
		block_version = block.header.version
		nonce = block.header.nonce
		bits = block.header.bits
		block_timestamp = block.header.time

		q1.put([current_block_hash,prev_block_hash,block_size,block_version,nonce,bits,block_timestamp])

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

				q3.put([txn_hash_outid,outid,output_val,address,txn_hash])

			inputs = transaction.inputs

			for i, inp in enumerate(inputs):

				prev_txn_hash = inp.previous_hash
				txn_hash_outid = prev_txn_hash+"_"+str(inp.txn_out_id)
				outid = inp.txn_out_id

				q4.put([txn_hash_outid,outid,txn_hash])


			q2.put([txn_hash,total_value,txn_version,txt_timestamp,current_block_hash])

		old_curr_hash = current_block_hash

	q5.put([file_no,block_no,c_blk_hash])


def listener(q1,q2,q3,q4,q5):

	block_csv = open('blocks.csv', 'a')
	transaction_csv = open('transactions.csv', 'a')
	output_csv = open('outputs.csv', 'a')
	in_payment_csv = open('in_payment.csv', 'a')
	last_ln = open("last_ln.csv",'a')

	block_fw = csv.writer(block_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
	transaction_fw = csv.writer(transaction_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
	output_fw = csv.writer(output_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
	in_payment_fw = csv.writer(in_payment_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
	last_ln_fw = csv.writer(last_ln, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)

	while True:

		m1 = q1.get()
		m2 = q2.get()
		m3 = q3.get()
		m4 = q4.get()
		m5 = q5.get()

		if m1 != "kill":
			block_fw.writerow(m1)
		else:
			done[0] = 0
		if m2 != "kill":
			transaction_fw.writerow(m2)
		else:
			done[1] = 0
		if m3 != "kill":
			output_fw.writerow(m3)
		else:
			done[2] = 0
		if m4 != "kill":
			in_payment_fw.writerow(m4)
		else:
			done[3] = 0
		if m5 != "kill":
			last_ln_fw.writerow(m5)
		else:
			done[4] = 0

		if sum(done) == 0:
			break

	block_csv.close()
	transaction_csv.close()
	output_csv.close()
	in_payment_csv.close()
	last_ln.close()

if __name__ == "__main__":

	headers = True
	done = [1,1,1,1,1]

	if headers:

		block_csv = open('blocks.csv', 'w')
		transaction_csv = open('transactions.csv', 'w')
		output_csv = open('outputs.csv', 'w')
		in_payment_csv = open('in_payment.csv', 'w')
		last_ln = open("last_ln.csv",'w')

		block_fw = csv.writer(block_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
		block_fw.writerow(["current_block_hash","prev_block_hash","block_size","block_version","nonce","bits","block_timestamp"])
		
		transaction_fw = csv.writer(transaction_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
		transaction_fw.writerow(["transaction_hash","total_value","version_no","timestamp","block_hash"])
		
		output_fw = csv.writer(output_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
		output_fw.writerow(["transaction_hash_outid","id","value","address","transaction_hash"])
		
		in_payment_fw = csv.writer(in_payment_csv, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
		in_payment_fw.writerow(["transaction_hash_outid","id","current_transaction_hash"])

		last_ln_fw = csv.writer(last_ln, delimiter=',',quotechar='|', quoting=csv.QUOTE_MINIMAL)
		last_ln_fw.writerow(["file_no","block_no","current_block_hash"])

		block_csv.close()
		transaction_csv.close()
		output_csv.close()
		in_payment_csv.close()
		last_ln.close()

	manager = mp.Manager()
	q1 = manager.Queue()
	q2 = manager.Queue()
	q3 = manager.Queue()
	q4 = manager.Queue()
	q5 = manager.Queue()
	pool = mp.Pool(mp.cpu_count() + 5)

	#put listener to work first
	watcher = pool.apply_async(listener, (q1,q2,q3,q4,q5))

	#fire off workers
	jobs = []
	for i in range(1360):
	    job = pool.apply_async(worker, (i, q1,q2,q3,q4,q5))
	    jobs.append(job)

	# collect results from the workers through the pool result queue
	for job in jobs: 
	    job.get()

	#now we are done, kill the listener
	q1.put('kill')
	q2.put('kill')
	q3.put('kill')
	q4.put('kill')
	q5.put('kill')
	pool.close()










    