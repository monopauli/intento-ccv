kill -9 $(lsof -t -i:26657 -sTCP:LISTEN)
kill -9 $(lsof -t -i:1317 -sTCP:LISTEN)
kill -9 $(lsof -t -i:6060 -sTCP:LISTEN)
kill -9 $(lsof -t -i:9090 -sTCP:LISTEN)
kill -9 $(lsof -t -i:54246  -sTCP:LISTEN)
kill -9 $(lsof -t -i:26656 -sTCP:LISTEN)
sudo apt remove trustless-hub-node -y
make clean-files
sudo apt install ./trustless-hub-node_0.7.5-9-gd2a970d_amd64.deb -y

./init.sh
