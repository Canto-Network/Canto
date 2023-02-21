// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

import "./turnstile.sol";
import "./csrSmartContract.sol";

contract FactoryContract {
    address turnstile;

    event Address(address smartContract);

    constructor(address _turnstile) {
        turnstile = _turnstile;
    }

    function register(address to) public returns (address) {
        CSRSmartContract c = new CSRSmartContract(turnstile);
        c.register(to);
        emit Address(address(c));
        return address(c);
    }

    function assign(uint256 id) public returns (address) {
        CSRSmartContract c = new CSRSmartContract(turnstile);
        c.assign(id);
        emit Address(address(c));
        return address(c);
    }
}