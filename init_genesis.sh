KEY="mykey"
KEY2="plexkey"
CHAINID="canto_740-1"
MONIKER="plex-validator"
KEYRING="os"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# Reinstall daemon
rm -rf ~/.cantod*
make install

# Set client config
cantod config keyring-backend $KEYRING
cantod config chain-id $CHAINID

# if $KEY exists it should be deleted
cantod keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO
cantod keys add $KEY2 --keyring-backend $KEYRING --algo $KEYALGO

# Set moniker and chain-id for Canto (Moniker can be anything, chain-id must be an integer)
cantod init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to acanto
cat $HOME/.cantod/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

# Change voting params so that submitted proposals pass immediately for testing
cat $HOME/.cantod/config/genesis.json| jq '.app_state.gov.voting_params.voting_period="70s"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json


# disable produce empty block
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.cantod/config/config.toml
  else
    sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.cantod/config/config.toml
fi

if [[ $1 == "pending" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.cantod/config/config.toml
      sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.cantod/config/config.toml
  else
      sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.cantod/config/config.toml
      sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.cantod/config/config.toml
  fi
fi

# Allocate genesis accounts (cosmos formatted addresses)
cantod add-genesis-account $KEY 850000000000000000000000000acanto --keyring-backend $KEYRING
cantod add-genesis-account $KEY2 35808383230000000000000000acanto --keyring-backend $KEYRING

# Contributors
cantod add-genesis-account canto18dlycyyudjw87sy6anzwl8xcsex65rx6mnjl4s 15568862280000000000000000acanto
cantod add-genesis-account canto1kej2wa54r3wva487qp6sd8hfgvydsme9k95ufp 15568862280000000000000000acanto
cantod add-genesis-account canto16xyggs3c49zf7vtmc42jr7zs8jc5cn2e05dxj7 12455089820000000000000000acanto
cantod add-genesis-account canto1jmns67pp5enklcrs9ekndrmxfwllzdl2kr6vhh 7784431138000000000000000acanto
cantod add-genesis-account canto1c7nla7zn0fkwz3rvlvlyljzlz2qywmtw5yruxw 4670658683000000000000000acanto
cantod add-genesis-account canto1pr3fp69cecefw5q7lt4uyvt4fpgmdk8gtny502 4670658683000000000000000acanto
cantod add-genesis-account canto1nw479jevfl8ql6t8fl4z5pdz22qpz3zpr9ej3n 1556886228000000000000000acanto
cantod add-genesis-account canto1g5ghweev32gdmmm2wkauecnvl7cgq7yh6ywea2 1556886228000000000000000acanto
cantod add-genesis-account canto1curufsg0j9a7qfxwl2ef7qxtcf4d7jh3nklyh9 1556886228000000000000000acanto
cantod add-genesis-account canto1qphfgwv8h594ma9q2jkmudghp9hypa3g66wf2f 1556886228000000000000000acanto
cantod add-genesis-account canto1x3apk8hcrtxhgaq0dcrn2ae502x4lug8fugsql 1556886228000000000000000acanto
cantod add-genesis-account canto16mrzhlxz4rr9vekjv7qye9rjr9uk3d8y9y00qm 778443113800000000000000acanto
cantod add-genesis-account canto1j2eg8p5zujqrs58hyasqwudsccctjn27758c43 778443113800000000000000acanto
cantod add-genesis-account canto1mlptkz63vk822vl8ekqeg9f80fmdsjxwzdwftp 778443113800000000000000acanto
cantod add-genesis-account canto17hugq2wjdfqkdhhps0v7v46fwzt8yfeltxe5mm 778443113800000000000000acanto
cantod add-genesis-account canto1g9wx5zgq2h5ekn30jr9x8y87fjf77e0rtuqw7c 778443113800000000000000acanto
cantod add-genesis-account canto18xvsay3mmz0u3748shzv5gnrkta47jqqtnpmpw 778443113800000000000000acanto
cantod add-genesis-account canto1j0x25v3le6qwn8edeczc6j25xny7usmug5226k 778443113800000000000000acanto
cantod add-genesis-account canto1twej433h8td4arrdh967w089c0x8gplv89tcap 311377245500000000000000acanto
cantod add-genesis-account canto1h249yu69yur9n4yn7czrmcn65a0pf8dk3hpyc5 389221556900000000000000acanto
cantod add-genesis-account canto18j2cgsnyxt4c28jk3y3s6h8u2pjezq74zx7ygs 77844311380000000000000acanto

# Settlers/Helpers


# Update total supply with claim values
#validators_supply=$(cat $HOME/.cantod/config/genesis.json | jq -r '.app_state["bank"]["supply"][0]["amount"]')
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
# total_supply=1000000000000000000000000000
# cat $HOME/.cantod/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

echo $KEYRING
echo $KEY
# Sign genesis transaction
cantod gentx $KEY2 100000000000000000000000acanto --keyring-backend $KEYRING --chain-id $CHAINID
#cantod gentx $KEY2 1000000000000000000000acanto --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
cantod collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
cantod validate-genesis

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
# cantod start --pruning=nothing --trace --log_level info --minimum-gas-prices=0.0001acanto --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable true

