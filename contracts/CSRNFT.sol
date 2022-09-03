// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {ERC721} from  "../lib/openzeppelin-contracts/contracts/token/ERC721/ERC721.sol";

contract CSRNFT is ERC721 {

    // error - not the withdrawer of nft
    error notWithdrawer(address sender, address actualWithdrawer, uint256 nftId);
    
    // error - not the csrModuleAcct
    error notCSRModuleAccount(address sender, address csrModuleAccount);

    // event to be emitted on withdrawals
    event Withdrawal(address withdrawer, address receiver, uint id);
    
    // hold the address of the CSR module account in state
    address public csrModuleAccount;
    // holds the total number of nfts minted
    uint256 public nfts;

    /** 
     * @param nftId which nftID is being withdrawn from
     * @param withdrawer account intending to withdraw rewards from CSR 
     attached to nftID
     */
    modifier checkOwnership(uint nftId, address withdrawer) {
        // fail if withdrawer is not the owner of this CSR
        if (super.ownerOf(nftId) != withdrawer) {
            revert notWithdrawer({
                sender:           withdrawer,
                actualWithdrawer: super.ownerOf(nftId), 
                nftId:            nftId
            });
        }
        _;
    }

    /**
     * @param sender the account sending the tx for minting
     * @dev minter is restricted to csrModuleAccount,
     */ 
    modifier isMinter(address sender) {
        if (sender != csrModuleAccount) {
            revert notCSRModuleAccount({
                sender:           sender,
                csrModuleAccount: csrModuleAccount
            });
        }
        _;
    }

    /**
     * @dev modify constructor so that the CSR module account is set as the minter 
     * for this contract.
     */
    constructor(string memory name_, string memory symbol_) ERC721(name_, symbol_) {
        // set csrModuleAccount to msg.sender, as the csr mod account will 
        // deploy this contract upon construction
        csrModuleAccount = msg.sender;
    }

    /** 
        * @param receiver, the account to receive withdrawn funds if suscessful 
        * @param nftId, the nft minted by the CSR being withdrawn from
    */
    function withdraw(address receiver, uint256 nftId) checkOwnership(nftId, msg.sender) public {
        // emit withdrawal method to be caught by CSR module
        emit Withdrawal({
            withdrawer: msg.sender,
            receiver:   receiver,
            id:         nftId
        });
      }

    /** 
      * @dev method to be called by CSR Module after creation of a new CSR
      * @param owner_, address to be minted an nft
      * @return nftId, the nftId minted to owner, registered as uint64
    */     
    function mintCSR(address owner_) isMinter(msg.sender) public returns(uint64) {
        // increment nft number
        nfts++;
        // mint this nft to the user
        super._safeMint(owner_, nfts, "");
    }
}

