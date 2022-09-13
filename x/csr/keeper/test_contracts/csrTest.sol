// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

import "./turnstile.sol";

contract CSRTest {
    event CreateEvent(string message, address sender);

    function register(address turnstile, address to) public {
        Turnstile(turnstile).register(to);
    }

    function update(address turnstile, uint256 nftID) public {
        Turnstile(turnstile).assign(nftID);
    }

    constructor() {
        emit CreateEvent("contract created", msg.sender);
    }
}
