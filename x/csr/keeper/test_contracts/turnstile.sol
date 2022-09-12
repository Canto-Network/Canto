// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

contract Turnstile {
    // UpdateEvent is emitted when a user wants to add new smart contracts
    // to the same cst NFT.
    event Attach(address smartContractAddress, uint64 id);
    // RegisterEvent is emitted when a user wants to create a new CSR nft
    event Register(address smartContractAddress, address receiver);
    // RetroactiveRegisterEvent is emitted when a user wants to retroactively register a smart
    // contract that was previously deployed
    event RetroactiveRegisterEvent(
        address[] contracts,
        address deployer,
        uint64[][] nonces,
        bytes[] salt,
        uint64 id
    );

    // register the smart contract to an existing CSR nft
    function register(uint64 id) public {
        emit Attach(msg.sender, id);
    }

    // register and mint a new CSR nft that will be transferred to the to address entered
    function register(address to) public {
        emit Register(msg.sender, to);
    }

    // deploys a smart contract and registers the newly deployed smart contract to
    // an existing CSR nft
    function deploy(bytes[] memory contractByteCode, uint64 nft_id) public {
        // createThisSmartContract(contractByteCode)
        // emit UpdateCSREvent(newSmartContractAddress, nft_id)
    }

    // deploys a smart contract and registers the newly deployed smart contract
    // to a new CSR nft
    function deploy(bytes[] memory contractByteCode) public {
        // createThisSmartContract(contractByteCode)
        // emit RegisterCSREvent(newSmartContractAddress, msg.sender)
    }

    // retroactively register a set of smart contracts to a particular csr nft
    function retroactiveRegister(
        address[] memory contracts,
        address deployer,
        uint64[][] memory nonces,
        bytes[] memory salt,
        uint64 id
    ) public {
        emit RetroactiveRegisterEvent(
            contracts,
            deployer,
            nonces,
            salt,
            id
        );
    }
}
