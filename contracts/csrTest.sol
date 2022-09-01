// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

import "./turnstile.sol";

contract CSRTest {
    event CreateEvent(string message, address sender);

    constructor(address turnstileContract) {
        Turnstile(turnstileContract).register(msg.sender);
        emit CreateEvent("contract created", msg.sender);
    }
}
