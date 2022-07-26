# GENTX INSTRUCTIONS

### Install & Initialize 

* Install cantod binary

* Initialize canto node directory 
```bash
cantod init <node_name> --chain-id <chain_id>
```
* Download the [genesis file](https://github.com/Canto-Network/Canto/raw/main/Mainnet/genesis.json)
```bash
wget https://github.com/Canto-Network/Canto/raw/main/Mainnet/genesis.json -b $HOME/.cantod/config
```

### Add a Genesis Account
A genesis account is required to create a GENTX

```bash
cantod add-genesis-account <address-or-key-name> acanto --chain-id <chain-id>
```
### Create & Submit a GENTX file + genesis.json
A GENTX is a genesis transaction that adds a validator node to the genesis file.
```bash
cantod gentx <key_name> <token-amount>acanto --chain-id=<chain_id> --moniker=<your_moniker> --commission-max-change-rate=0.01 --commission-max-rate=0.10 --commission-rate=0.05 --details="<details here>" --security-contact="<email>" --website="<website>"
```
* Fork [Canto](https://github.com/Canto-Network/Canto)

* Copy the contents of `${HOME}/.cantod/config/gentx/gentx-XXXXXXXX.json` to `$HOME/Canto/Mainnet/gentx/<yourvalidatorname>.json`

* Copy the genesis.json file `${HOME}/.cantod/config/genesis.json` to `$HOME/Canto/Mainnet/Genesis-Files/`

* Create a pull request to the main branch of the [repository](https://github.com/Canto-Network/Canto/Mainnet/gentx)