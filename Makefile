run:
	API_ADDR=localhost:8080 \
	UDP_DIAL_ADDR=localhost:7777 \
	UDP_LISTEN_ADDR=127.0.0.1:7777 \
	BROADCAST_FREQ_HZ=1000 \
	go run main.go