// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Test} from "forge-std/Test.sol";
import {CEXToken} from "../src/token/CEXToken.sol";

contract CEXTokenTest is Test {
    CEXToken token;
    address minter = address(0xM1);
    address alice = address(0xA1);
    address bob = address(0xB1);

    function setUp() public {
        token = new CEXToken("Mock USDT", "USDT", 6, minter);
    }

    function test_MintAndTransfer() public {
        vm.prank(minter);
        token.mint(alice, 1_000_000);

        assertEq(token.balanceOf(alice), 1_000_000);

        vm.prank(alice);
        token.transfer(bob, 250_000);

        assertEq(token.balanceOf(alice), 750_000);
        assertEq(token.balanceOf(bob), 250_000);
    }

    function test_OnlyMinterCanMint() public {
        vm.prank(alice);
        vm.expectRevert(CEXToken.Unauthorized.selector);
        token.mint(alice, 100);
    }
}
