// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

import "./turnstile.sol";

contract CSRTest {
    event CreateEvent(string message, address sender);

    function register(address turnstile, address to) public {
        Turnstile(turnstile).register(to);
    }

    constructor() {
        emit CreateEvent("contract created", msg.sender);
    }
}
