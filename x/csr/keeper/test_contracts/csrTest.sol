// SPDX-License-Identifier: MIT

pragma solidity ^0.8.16;

import "../../../../contracts/turnstile.sol";

contract CSRTest {
    event CreateEvent(string message, address sender);

    function register(address turnstile) public {
        Turnstile(turnstile).register(msg.sender);
    }

    constructor() {
        emit CreateEvent("contract created", msg.sender);
    }
}
