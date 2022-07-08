FROM ubuntu:latest 

# add your keyname
ARG KEY_NAME <add your keyname>

# add the nickname for your validator on the network
ARG MONIKER <add your moniker>

# add the network id
ARG NETWORK_ID canto_9624-1

# arbitrary validator metadata
ARG DETAILS <details as string>

# add amount you're delegating
ARG TOKEN_DELEGATION <amount>

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update &&\ 
	apt-get install -y apt-utils git git-extras software-properties-common vim \
	wget ca-certificates curl build-essential libssl-dev make openssh-server snap

RUN snap install go --classic

# builds out cantod
RUN git clone https://github.com/Canto-Network/Canto-Testnet
WORKDIR /Canto-Testnet/cmd/cantod
RUN go install -tags ledger ./... &&\
    mv $HOME/go/bin/cantod /usr/bin/

# generate keys
RUN cantod keys add $KEY_NAME
    # cantod keys add $KEY_NAME --recover &&\
    # cantod keys add $KEY_NAME --ledger 

# initialize node
RUN cantod init $MONIKER --chain-id canto_9624-1

# get genesis file
RUN wget https://github.com/Canto-Network/Canto-Testnet-v2/raw/main/genesis.json -P $HOME/.cantod/config/
RUN sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001acanto"/g' $HOME/.cantod/config/app.toml
RUN wget https://github.com/Canto-Network/Canto-Testnet-v2/raw/main/Docker/cantod.service -P $HOME/systemd/system/

# run cantod
WORKDIR /Canto-Testnet/cmd/cantod
RUN sudo systemctl daemon-reload &&\
    sudo systemctl enable cantod.service &&\
    systemctl start cantod && journalctl -u cantod -f

ENTRYPOINT ["/bin/bash"]
