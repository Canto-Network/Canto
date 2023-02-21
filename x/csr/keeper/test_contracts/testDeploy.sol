pragma solidity ^0.8.16;

contract testDeploy {

    // number for querying, incremented each call
    uint64  private number = 0;

    //  child address of testDeploy Contract
    address private childAddress;
    // bytecode for child contract
    bytes private bytecode;

    constructor(bytes memory bytecode_) {
        bytecode = bytecode_;
    }

    // function to retrieve data
    function getNumber() public returns(uint) {
        number++;
        return number;
    }

    // function to return address derived from CREATE2 
    function deploy1 () public returns(address) {
        // load bytecode into memory 
        bytes memory code = bytecode;
        address addr;
        assembly {
            //  deploy contract
            addr := create(0, add(code, 32), mload(code))
        }
        return addr;
    }

    // function to return address derived from CREATE1 
    function deploy2(bytes32 salt_) public returns(address) {
        // load bytecode into memory
        bytes memory code = bytecode;
        address addr;
        assembly {
            // value for contract deployments
            let v := 0
            // retrieve code size at this address, first 32 bytes are the length of the
            let p := add(code, 32)
            // retrieve bytecode from memory 
            addr := create2(0, p, mload(code), salt_)
        }
        return addr;
    }
}