# Experiment Commands

### one-download

```
sudo tcpdump -i port1 -w one-download.pcap
./game -client -addr 10.0.0.1:8888 -t 40 -tick 50 -tag one-download

# wait for 5 sec -- to capture baseline
./download -client -addr 10.0.1.1:9999 -filesize 80000000

pigz one-download.pcap
```

### 100-downloads-n-flows

```
sudo tcpdump -i port1 -w 100-downloads-1-flow.pcap
./game -client -addr 10.0.0.1:8888 -t 40 -tick 50 -tag 100-downloads-1-flow

# wait for 5 sec -- to capture baseline
./download -client -addr 10.0.1.1:9999 -filesize 800000 -c 100 -p <n>

pigz 100-downloads.pcap
```

### 1-download-1-monitor

```
sudo tcpdump -i port1 -w 1-download-1-monitor.pcap
./game -client -addr 10.0.0.1:8888 -t 45 -tick 50 -tag 1-download-1-monitor

# wait for 5 sec -- to capture baseline
./download -client -addr 10.0.1.1:9999 -filesize 80000000

# start monitor flow in another docker exec into download_client
./download -client -addr 10.0.1.1:9999 -filesize 400000 -c 100

pigz 1-download-1-monitor.pcap
```