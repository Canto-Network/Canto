// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import "forge-std/Test.sol";
import "../src/CSRNFT.sol";

contract CSRNFTTest is Test {
    CSRNFT public csr;
    // admin for deploying contract, replicates functionality of csrModuleAcct
    address public admin = address(1);

    function setUp() public {
        vm.prank(admin);
        csr = new CSRNFT("nft", "nft");
    }

    // testing ownership, test that csrModuleAccount is set to admin
    function testOwnership() public {
        assertEq(csr.csrModuleAccount(), admin);
    }

    // test minting csrs, 
    function testMinting() public {
        assertEq(csr.nfts(), 0);
        vm.prank(admin);
        csr.mintCSR(admin);
        assertEq(csr.ownerOf(1), admin);
        assertEq(csr.nfts(), 1);
    }

    // test fails bc sender is not admin
    function testFailMinting() public {
        csr.mintCSR(admin);
    }
    event Withdrawal(address withdrawer, address receiver, uint id);

    // test withdrawal event emission
    function testWithdraw() public {
        vm.prank(admin);
        csr.mintCSR(admin);

        vm.expectEmit(true, true, true, true, address(csr));
        emit Withdrawal(admin, admin, 1);

        vm.prank(admin);
        csr.withdraw(admin, 1);
    }
}
