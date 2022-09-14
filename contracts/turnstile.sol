// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

contract Turnstile {
    uint256 private tokenID;
    uint256 private balance;
    mapping(uint256 => uint256) private distributions;

    // Attach is emitted when a user wants to add new smart contracts
    // to the same CSR NFT.
    event Attach(address smartContractAddress, uint256 id);
    // RegisterEvent is emitted when a user wants to create a new CSR nft
    event Register(address smartContractAddress, address receiver, uint256 id);

    constructor() {
        tokenID = 0;
    }

    // register the smart contract to an existing CSR nft
    function assign(uint256 id) public {
        tokenID++;
        emit Attach(msg.sender, id);
    }

    // register and mint a new CSR nft that will be transferred to the to address entered
    function register(address to) public {
        tokenID++;
        emit Register(msg.sender, to, tokenID);
    }

    function distributeFees(uint256 _tokenID) public payable {
        balance += msg.value;
        distributions[_tokenID] += msg.value;
    }

    function numberNFTs() public view returns (uint256) {
        return tokenID;
    }

    function turnstileBalance() public view returns (uint256) {
        return balance;
    }

    function revenue(uint256 _tokenID) public view returns (uint256) {
        return distributions[_tokenID];
    }
}
