#!/bin/bash

# Create screen session and run commands
# Bootnode session
screen -dmS bootnode
screen -S bootnode -X stuff 'cd /root/geth-poa-tutorial/node2 && geth --datadir ./data --syncmode 'full' --port 30312 --rpc --rpcaddr '0.0.0.0' --rpccorsdomain "*" --rpcport 8503 --rpcapi 'personal,db,eth,net,web3,txpool,miner'  --networkid 1515 --gasprice '1' --allow-insecure-unlock -unlock 0xFe2a7e374320Abe858c21310E533E169236e0F7E --password password.txt console --rpcvhosts='*' --rpccorsdomain='*' --gcmode "archive" --miner.etherbase 0xFe2a7e374320Abe858c21310E533E169236e0F7E \n'

# Watcher session
screen -dmS watcher
screen -S watcher -X stuff 'cd /var/www/zkScam && conda activate web3 && python server_Watcher.py\n'

# Sync session
screen -dmS sync
screen -S sync -X stuff 'cd /var/www/zkScam && conda activate web3 && python server_sync.py\n'

# Explorer session
screen -dmS explorer
screen -S explorer -X stuff 'cd /var/www/zkScam && conda activate web3 && python server_explorer.py\n'


