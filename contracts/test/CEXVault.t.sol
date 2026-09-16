// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Test} from "forge-std/Test.sol";
import {CEXToken} from "../src/token/CEXToken.sol";
import {Treasury} from "../src/treasury/Treasury.sol";
import {CEXVault} from "../src/exchange/CEXVault.sol";

contract CEXVaultTest is Test {
    CEXToken usdt;
    Treasury treasury;
    CEXVault vault;

    address operator = address(0xOP);
    address alice = address(0xA1);

    function setUp() public {
        usdt = new CEXToken("Mock USDT", "USDT", 6, address(this));
        treasury = new Treasury(operator);
        vault = new CEXVault(operator, address(treasury));

        vm.prank(operator);
        vault.setSupportedToken(address(usdt), true);

        usdt.mint(alice, 1_000_000);
    }

    function test_DepositEmitsEvent() public {
        vm.startPrank(alice);
        usdt.approve(address(vault), 500_000);

        vm.expectEmit(true, true, false, true);
        emit CEXVault.Deposit(alice, address(usdt), 500_000, block.timestamp);
        vault.deposit(address(usdt), 500_000);
        vm.stopPrank();

        assertEq(usdt.balanceOf(address(vault)), 500_000);
    }

    function test_OperatorWithdrawal() public {
        vm.startPrank(alice);
        usdt.approve(address(vault), 300_000);
        vault.deposit(address(usdt), 300_000);
        vm.stopPrank();

        address recipient = address(0xR1);
        bytes32 withdrawalId = keccak256("wd-1");

        vm.prank(operator);
        vault.withdraw(recipient, address(usdt), 200_000, withdrawalId);

        assertEq(usdt.balanceOf(recipient), 200_000);
        assertEq(usdt.balanceOf(address(vault)), 100_000);
    }

    function test_WithdrawalIdempotency() public {
        vm.startPrank(alice);
        usdt.approve(address(vault), 100_000);
        vault.deposit(address(usdt), 100_000);
        vm.stopPrank();

        bytes32 withdrawalId = keccak256("wd-dup");

        vm.startPrank(operator);
        vault.withdraw(alice, address(usdt), 50_000, withdrawalId);
        vm.expectRevert(CEXVault.WithdrawalAlreadyProcessed.selector);
        vault.withdraw(alice, address(usdt), 50_000, withdrawalId);
        vm.stopPrank();
    }

    function test_SweepFeesToTreasury() public {
        vm.startPrank(alice);
        usdt.approve(address(vault), 100_000);
        vault.deposit(address(usdt), 100_000);
        vm.stopPrank();

        vm.prank(operator);
        vault.sweepFees(address(usdt), 10_000);

        assertEq(usdt.balanceOf(address(treasury)), 10_000);
        assertEq(usdt.balanceOf(address(vault)), 90_000);
    }

    function test_NonOperatorCannotWithdraw() public {
        vm.prank(alice);
        vm.expectRevert(CEXVault.Unauthorized.selector);
        vault.withdraw(alice, address(usdt), 1, keccak256("x"));
    }
}
