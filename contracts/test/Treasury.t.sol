// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Test} from "forge-std/Test.sol";
import {CEXToken} from "../src/token/CEXToken.sol";
import {Treasury} from "../src/treasury/Treasury.sol";

contract TreasuryTest is Test {
    CEXToken token;
    Treasury treasury;
    address operator = address(0xOP);
    address vault = address(0xV1);

    function setUp() public {
        token = new CEXToken("Fee Token", "FEE", 18, address(this));
        treasury = new Treasury(operator);
        token.mint(vault, 1_000);
    }

    function test_ReceiveFees() public {
        vm.startPrank(vault);
        token.approve(address(treasury), 100);
        treasury.receiveFees(address(token), 100);
        vm.stopPrank();

        assertEq(token.balanceOf(address(treasury)), 100);
    }

    function test_OperatorWithdraw() public {
        vm.startPrank(vault);
        token.approve(address(treasury), 200);
        treasury.receiveFees(address(token), 200);
        vm.stopPrank();

        address cold = address(0xC1);
        vm.prank(operator);
        treasury.withdraw(cold, address(token), 150);

        assertEq(token.balanceOf(cold), 150);
    }
}
