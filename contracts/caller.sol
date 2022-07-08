pragma solidity ^0.8.10;

import "./Proposal-Store.sol";


// caller contract, passed a reference of the Proposal-Store contract
contract caller {
    address public unigovContract;

    function setPropContract(address propContract) external {
        require(unigovContract == address(0));
        unigovContract = propContract;
    }

    function queryProp(uint id) external {
        ProposalStore propStore = ProposalStore(unigovContract);
        ProposalStore.Proposal memory x = propStore.QueryProp(id);
        //Query this proposal and ping the callee contract
        bytes memory calldatas = abi.encodePacked(bytes4(keccak256(bytes(x.signatures[0]))), x.calldatas[0]);
        (bool success, bytes memory data) = x.targets[0].call{value: x.values[0]}(calldatas);
        //in this case we are pinging the setInt of the callee ... 
    }
}