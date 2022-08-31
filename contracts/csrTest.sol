pragma solidity ^0.8.16;

contract Token {
    event RandomEvent(string message, address sender);

    function emitEvent() public {
        emit RandomEvent("updated event", msg.sender);
    }

    constructor() {
        emitEvent();
    }
}
