// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

import "./turnstile.sol";

contract CSRSmartContract {
    address turnstile;

    constructor(address _turnstile) {
        turnstile = _turnstile;
    }

    function register(address to) public {
        Turnstile(turnstile).register(to);
    }

    function assign(uint256 id) public {
        Turnstile(turnstile).assign(id);
    }
}
