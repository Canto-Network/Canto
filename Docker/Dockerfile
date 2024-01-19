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

# builds out razord
RUN git clone https://github.com/Canto-Network/Canto-Testnet
WORKDIR /Canto-Testnet/cmd/razord
RUN go install -tags ledger ./... &&\
    mv $HOME/go/bin/razord /usr/bin/

# generate keys
RUN razord keys add $KEY_NAME
    # razord keys add $KEY_NAME --recover &&\
    # razord keys add $KEY_NAME --ledger 

# initialize node
RUN razord init $MONIKER --chain-id canto_9624-1

# get genesis file
RUN wget https://github.com/Canto-Network/Canto-Testnet-v2/raw/main/genesis.json -P $HOME/.razord/config/
RUN sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001acanto"/g' $HOME/.razord/config/app.toml
RUN wget https://github.com/Canto-Network/Canto-Testnet-v2/raw/main/Docker/razord.service -P $HOME/systemd/system/

# run razord
WORKDIR /Canto-Testnet/cmd/razord
RUN sudo systemctl daemon-reload &&\
    sudo systemctl enable razord.service &&\
    systemctl start razord && journalctl -u razord -f

ENTRYPOINT ["/bin/bash"]
