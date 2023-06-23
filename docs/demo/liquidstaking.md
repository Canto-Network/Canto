# Demo

## Changelog

## Demo

In this demo, you will learn how to use the liquid staking module.

### Build from source

```bash
# Clone the demo project and build `crescentd` for testing
git clone https://github.com/b-harvest/Canto
cd Canto
git checkout liquidstaking-module
make install
```

### Spin up a local node

* Init
  * [init_testnet.sh](https://github.com/b-harvest/Canto/blob/liquidstaking-module/init_testnet.sh)

* Run a node
  * `cantod start --pruning=nothing --trace --log_level trace --minimum-gas-prices=1.000acanto --json-rpc.api eth,txpool,personal,net,debug,web3 --rpc.laddr "tcp://0.0.0.0:26657" --api.enable true --api.enabled-unsafe-cors true`

### Provide Insurance
```bash
# Check the validator address first
cantod query staking validators

# Provide insurances with the validator_address we queried above (fee: 10%)
cantod tx liquidstaking provide-insurance <validator_address> 17500000000000000000000acanto 0.1 --from key1 --fees 200000acanto
cantod tx liquidstaking provide-insurance <validator_address> 17500000000000000000000acanto 0.1 --from key1 --fees 200000acanto
cantod tx liquidstaking provide-insurance <validator_address> 17500000000000000000000acanto 0.1 --from key1 --fees 200000acanto

# You can query the insurance you provide
cantod query liquidstaking insurances
```

### Liquid Stake
```bash
# Now we have pairing insurance, so we can liquid stake
# Liquid stake 3 chunks
cantod tx liquidstaking liquid-stake 250000000000000000000000acanto --from key1 --fees 3000000acanto --gas 3000000
cantod tx liquidstaking liquid-stake 250000000000000000000000acanto --from key1 --fees 3000000acanto --gas 3000000
cantod tx liquidstaking liquid-stake 250000000000000000000000acanto --from key1 --fees 3000000acanto --gas 3000000

# You can query the insurance you provide
cantod query liquidstaking chunks

# Check states to see if the liquid staking is successful
cantod query states
```

### Liquid Unstake
```bash
# Request unstake
cantod tx liquidstaking liquid-unstake 250000000000000000000000acanto --from key1 --fees 3000000acanto --gas 3000000
# Unstake is not started yet, so we can query the unstake request
cantod query liquidstaking unpairing-for-unstaking-chunk-infos --queued=true
```

### Withdraw insurance
```bash
cantod tx liquidstaking withdraw-insurance 1 --from key1 --fees 3000000acanto --gas 3000000
# Query to see request is successfully queued or not
# Queued means the request will be handled at epoch.
cantod query liquidstaking withdraw-insurance-requests
```

### Query Epoch
```bash
# You can query the epoch so that you can check when queued liquid unstakes and withdraw insurances will be handled
cantod query liquidstaking epoch
```