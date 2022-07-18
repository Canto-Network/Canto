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

**If using Ubuntu:**

Install all dependencies:

`sudo snap install go --classic && sudo apt-get install git && sudo apt-get install gcc && sudo apt-get install make`

Or install individually:

* go1.18+: `sudo snap install go --classic`
* git: `sudo apt-get install git`
* gcc: `sudo apt-get install gcc`
* make: `sudo apt-get install make`

**If using Arch Linux:**

* go1.18+: `pacman -S go`
* git: `pacman -S git`
* gcc: `pacman -S gcc`
* make: `pacman -S make`

## Install `cantod`

### Clone git repository

```bash
git clone https://github.com/Canto-Network/Canto.git
cd Canto/cmd/cantod
go install -tags ledger ./...
sudo mv $HOME/go/bin/cantod /usr/bin/

```

### Generate and store keys

Replace `<keyname>` below with whatever you'd like to name your key.

*  `cantod keys add <key_name>`
*  `cantod keys add <key_name> --recover` to regenerate keys with your mnemonic
*  `cantod keys add <key_name> --ledger` to generate keys with ledger device

Store a backup of your keys and mnemonic securely offline.

Then save the generated public key config in the main Canto directory as `<key_name>.info`. It should look like this:

```

pubkey: {
  "@type":" ethermint.crypto.v1.ethsecp256k1.PubKey",
  "key":"############################################"
}

```

You'll use this file later when creating your validator txn.

## Set up validator

Install cantod binary from `Canto` directory: 

`sudo make install`

Initialize the node. Replace `<moniker>` with whatever you'd like to name your validator.

`cantod init <moniker> --chain-id canto_7744-1`

If this runs successfully, it should dump a blob of JSON to the terminal.

Download the Genesis file: 

`wget https://github.com/Canto-Network/Canto/raw/main/Mainnet/genesis.json -P $HOME/.cantod/config/` 

> _**Note:** If you later get `Error: couldn't read GenesisDoc file: open /root/.cantod/config/genesis.json: no such file or directory` put the genesis.json file wherever it wants instead, such as:
> 
> `sudo wget https://github.com/Canto-Network/Canto/raw/main/Mainnet/genesis.json -P/root/.cantod/config/`

Edit the minimum-gas-prices in `${HOME}/.cantod/config/app.toml`:

`sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001acanto"/g' $HOME/.cantod/config/app.toml`

### Set `cantod` to run automatically

* Start `cantod` by creating a systemd service to run the node in the background: 
* Edit the file: `nano /etc/systemd/system/cantod.service`
* Then copy and paste the following text into your service file. Be sure to edit as you see fit.

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

Reload the service files: 

`sudo systemctl daemon-reload`

Create the symlinlk: 

`sudo systemctl enable cantod.service`

Start the node: 

`sudo systemctl start cantod && journalctl -u cantod -f`

### Create Validator Transaction

Modify the following items below, removing the `<>`

- `<KEY_FILENAME>` should be `canto-keys` if you followed the steps above, but if you saved your public key JSON file above by any name but `canto-keys.info` you'll want to use that name here.
- `<VALIDATOR_NAME>` is whatever you'd like to name your node
- `<DESCRIPTION>` is whatever you'd like in the description field for your node
- `<SECURITY_CONTACT_EMAIL>` is the email you want to use in the event of a security incident
- `<YOUR_WEBSITE>` the website you want associated with your node
- `<TOKEN_DELEGATION>` is the amount of tokens staked by your node


```bash

cantod tx staking create-validator \
--from <KEY_FILENAME> \
--chain-id canto_7744-1 \
--moniker="<VALIDATOR_NAME>" \
--commission-max-change-rate=0.01 \
--commission-max-rate=1.0 \
--commission-rate=0.05 \
--details="<DESCRIPTION>" \
--security-contact="<SECURITY_CONTACT_EMAIL>" \
--website="<YOUR_WEBSITE>" \
--pubkey $(cantod tendermint show-validator) \
--min-self-delegation="1" \
--amount <TOKEN_DELEGATION>acanto \
--node http://164.90.134.106:26657 \
--fees 20acanto

```