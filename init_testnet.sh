KEY="mykey"
KEY2="mykey2"
KEY3="mykey3"
CHAINID="canto_7722-1"
MONIKER="plex-validator"
KEYRING="test"
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
cantod keys add $KEY3 --keyring-backend $KEYRING --algo $KEYALGO

# Set moniker and chain-id for Canto (Moniker can be anything, chain-id must be an integer)
cantod init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to acanto
cat $HOME/.cantod/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="acanto"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json
cat $HOME/.cantod/config/genesis.json | jq '.app_state["inflation"]["params"]["enable_inflation"]=false' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json


# Change voting params so that submitted proposals pass immediately for testing
cat $HOME/.cantod/config/genesis.json| jq '.app_state.gov.voting_params.voting_period="120s"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

cat $HOME/.cantod/config/genesis.json| jq '.app_state.gov.tally_params.threshold="0.000000000000000000"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

cat $HOME/.cantod/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["amount"]="1"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json


cat $HOME/.cantod/config/genesis.json| jq '.app_state.gov.tally_params.veto_threshold="0.000000000000000000"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json


cat $HOME/.cantod/config/genesis.json| jq '.app_state.gov.tally_params.quorum="0.000000000000000000"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

# Set gas limit in genesis
cat $HOME/.cantod/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="10000000"' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

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
cantod add-genesis-account $KEY 964723926400000000000000000acanto --keyring-backend $KEYRING
cantod add-genesis-account $KEY2 35276073620000000000000000acanto --keyring-backend $KEYRING
                                 
# Update total supply with claim values
#validators_supply=$(cat $HOME/.cantod/config/genesis.json | jq -r '.app_state["bank"]["supply"][0]["amount"]')
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
total_supply=1000000000000000000000000000
cat $HOME/.cantod/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.cantod/config/tmp_genesis.json && mv $HOME/.cantod/config/tmp_genesis.json $HOME/.cantod/config/genesis.json

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
cantod start --pruning=nothing --trace --log_level info --minimum-gas-prices=0.0001acanto --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable true

