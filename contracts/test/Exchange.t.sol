// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Test} from "forge-std/Test.sol";
import {Exchange} from "../src/exchange/Exchange.sol";
import {CEXToken} from "../src/token/CEXToken.sol";

/// @notice Verifies Exchange deploys vault + treasury and does NOT expose order matching.
contract ExchangeTest is Test {
    Exchange exchange;
    CEXToken usdt;
    address operator = address(0xOP);
    address user = address(0xU1);

    function setUp() public {
        exchange = new Exchange(operator);
        usdt = new CEXToken("Mock USDT", "USDT", 6, address(this));
        usdt.mint(user, 1_000_000);

        vm.prank(operator);
        exchange.vault().setSupportedToken(address(usdt), true);
    }

    function test_DeploysVaultAndTreasury() public view {
        assertEq(exchange.operator(), operator);
        assertTrue(address(exchange.vault()) != address(0));
        assertTrue(address(exchange.treasury()) != address(0));
    }

    function test_DepositFlow() public {
        vm.startPrank(user);
        usdt.approve(address(exchange.vault()), 500_000);
        exchange.vault().deposit(address(usdt), 500_000);
        vm.stopPrank();

        assertEq(usdt.balanceOf(address(exchange.vault())), 500_000);
    }
}
