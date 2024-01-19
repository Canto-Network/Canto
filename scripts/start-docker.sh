#!/bin/bash

KEY="mykey"
CHAINID="canto_9000-1"
MONIKER="mymoniker"
DATA_DIR=$(mktemp -d -t canto-datadir.XXXXX)

echo "create and add new keys"
./razord keys add $KEY --home $DATA_DIR --no-backup --chain-id $CHAINID --algo "eth_secp256k1" --keyring-backend test
echo "init canto with moniker=$MONIKER and chain-id=$CHAINID"
./razord init $MONIKER --chain-id $CHAINID --home $DATA_DIR
echo "prepare genesis: Allocate genesis accounts"
./razord add-genesis-account \
"$(./razord keys show $KEY -a --home $DATA_DIR --keyring-backend test)" 1000000000000000000acanto,1000000000000000000stake \
--home $DATA_DIR --keyring-backend test
echo "prepare genesis: Sign genesis transaction"
./razord gentx $KEY 1000000000000000000stake --keyring-backend test --home $DATA_DIR --keyring-backend test --chain-id $CHAINID
echo "prepare genesis: Collect genesis tx"
./razord collect-gentxs --home $DATA_DIR
echo "prepare genesis: Run validate-genesis to ensure everything worked and that the genesis file is setup correctly"
./razord validate-genesis --home $DATA_DIR

echo "starting canto node $i in background ..."
./razord start --pruning=nothing --rpc.unsafe \
--keyring-backend test --home $DATA_DIR \
>$DATA_DIR/node.log 2>&1 & disown

echo "started canto node"
tail -f /dev/null