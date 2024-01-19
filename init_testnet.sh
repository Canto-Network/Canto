KEY1="key1"
KEY2="key2"
KEY3="key3"
CHAINID="razor_777-1 "
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
rm -rf ~/.razord*
make install-testing

# Set client config
razord config keyring-backend $KEYRING
razord config chain-id $CHAINID

# if $KEY exists it should be deleted
razord keys add $KEY1 --keyring-backend $KEYRING --algo $KEYALGO
razord keys add $KEY2 --keyring-backend $KEYRING --algo $KEYALGO
razord keys add $KEY3 --keyring-backend $KEYRING --algo $KEYALGO



# Set moniker and chain-id for Canto (Moniker can be anything, chain-id must be an integer)
razord init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to arazor
cat $HOME/.razord/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="arazor"' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json
cat $HOME/.razord/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="arazor"' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json
cat $HOME/.razord/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="arazor"' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json
cat $HOME/.razord/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="arazor"' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json
cat $HOME/.razord/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="arazor"' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json

# Change voting params so that submitted proposals pass immediately for testing
cat $HOME/.razord/config/genesis.json| jq '.app_state.gov.voting_params.voting_period="30s"' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json


# disable produce empty block
if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.razord/config/config.toml
  else
    sed -i 's/create_empty_blocks = true/create_empty_blocks = false/g' $HOME/.razord/config/config.toml
fi

if [[ $1 == "pending" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.razord/config/config.toml
      sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.razord/config/config.toml
  else
      sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' $HOME/.razord/config/config.toml
      sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' $HOME/.razord/config/config.toml
  fi
fi

# Allocate genesis accounts (cosmos formatted addresses)
razord add-genesis-account $KEY1 105000000000000000000000000arazor --keyring-backend $KEYRING
razord add-genesis-account $KEY2 100000000000000000000000000arazor --keyring-backend $KEYRING
razord add-genesis-account $KEY3 100000000000000000000000000arazor --keyring-backend $KEYRING



# Update total supply with claim values
#validators_supply=$(cat $HOME/.razord/config/genesis.json | jq -r '.app_state["bank"]["supply"][0]["amount"]')
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$amount_to_claim+$validators_supply")
total_supply=305000000000000000000000000
cat $HOME/.razord/config/genesis.json | jq -r --arg total_supply "$total_supply" '.app_state["bank"]["supply"][0]["amount"]=$total_supply' > $HOME/.razord/config/tmp_genesis.json && mv $HOME/.razord/config/tmp_genesis.json $HOME/.razord/config/genesis.json

echo $KEYRING
echo $KEY1
# Sign genesis transaction
mkdir $HOME/.razord/config/gentx
razord gentx $KEY1 900000000000000000000000arazor --keyring-backend $KEYRING --chain-id $CHAINID --output-document $HOME/.razord/config/gentx/gentx-1.json
#razord gentx $KEY2 1000000000000000000000arazor --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
razord collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
razord validate-genesis

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
#razord start --pruning=nothing --trace --log_level trace --minimum-gas-prices=1.000arazor --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable true --api.enabled-unsafe-cors true

