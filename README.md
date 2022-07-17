# Becoming A Validator

**How to validate on the Canto Mainnet**

*(canto_7744-1)*

> Genesis file [Published](https://github.com/Canto-Network/Canto/raw/main/Mainnet/genesis.json)
> Peers list [Published](https://github.com/Canto-Network/Canto/blob/main/Mainnet/peers.txt)

## Hardware Requirements

### Minimum:
* 16 GB RAM
* 100 GB NVME SSD
* 3.2 GHz x4 CPU

### Recommended:
* 32 GB RAM
* 500 GB NVME SSD
* 4.2 GHz x6 CPU

### Operating System:
* Linux (x86_64) or Linux (amd64)
* Recommended Ubuntu or Arch Linux

## Install dependencies 

If using Ubuntu:

* go1.18+: `sudo snap install go --classic`
* git: `sudo apt-get install git`
* gcc: `sudo apt-get install gcc`
* make: `sudo apt-get install make`

If using Arch Linux:

* go1.18+: `pacman -S go`
* git: `pacman -S git`
* gcc: `pacman -S gcc`
* make: `pacman -S make`

## Cantod Installation Steps

**Clone git repository**

```bash
git clone https://github.com/Canto-Network/Canto.git
cd Canto/cmd/cantod
go install -tags ledger ./...
mv $HOME/go/bin/cantod /usr/bin/

```

**Generate keys**

*  `cantod keys add [key_name]`
*  `cantod keys add [key_name] --recover` to regenerate keys with your mnemonic
*  `cantod keys add [key_name] --ledger` to generate keys with ledger device

## Validator setup instructions

* Install cantod binary
* Initialize node: `cantod init <moniker> --chain-id canto_7744-1`
* Download the Genesis file: `wget https://github.com/Canto-Network/Canto/raw/main/Mainnet/genesis.json -P $HOME/.cantod/config/` (be sure to `rm -rf genesis.json` as it is automatically generated upon init.) 
* Edit the minimum-gas-prices in ${HOME}/.cantod/config/app.toml: `sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001acanto"/g' $HOME/.cantod/config/app.toml`
* Start cantod by creating a systemd service to run the node in the background

`nano /etc/systemd/system/cantod.service`

> Copy and paste the following text into your service file. Be sure to edit as you see fit.

```bash

[Unit]
Description=Canto Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/
ExecStart=/root/go/bin/cantod start --trace --log_level info --json-rpc.api eth,txpool,personal,net,debug,web3 --api.enable
Restart=on-failure
StartLimitInterval=0
RestartSec=3
LimitNOFILE=65535
LimitMEMLOCK=209715200

[Install]
WantedBy=multi-user.target

```

## Start the node

* Reload the service files: `sudo systemctl daemon-reload`
* Create the symlinlk: `sudo systemctl enable cantod.service`
* Start the node sudo: `systemctl start cantod && journalctl -u cantod -f`

### Create Validator Transaction

```bash

cantod tx staking create-validator \
--from {{KEY_NAME}} \
--chain-id canto_7744-1 \
--moniker="<VALIDATOR_NAME>" \
--commission-max-change-rate=0.01 \
--commission-max-rate=1.0 \
--commission-rate=0.05 \
--details="<description>" \
--security-contact="<contact_information>" \
--website="<your_website>" \
--pubkey $(cantod tendermint show-validator) \
--min-self-delegation="1" \
--amount <token delegation>acanto \
--node http://164.90.134.106:26657 \
--fees 20acanto

```