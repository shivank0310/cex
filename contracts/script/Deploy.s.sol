// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Script} from "forge-std/Script.sol";
import {Exchange} from "../src/exchange/Exchange.sol";
import {CEXToken} from "../src/token/CEXToken.sol";

/// @notice Deploy CEX on-chain contracts for local/testnet.
/// @dev Run: forge script script/Deploy.s.sol --rpc-url <url> --broadcast
contract Deploy is Script {
    function run() external {
        address operator = vm.envOr("CEX_OPERATOR", address(0xBEEF));

        vm.startBroadcast();
        Exchange exchange = new Exchange(operator);

        CEXToken usdt = new CEXToken("CEX USDT", "USDT", 6, operator);
        CEXToken wbtc = new CEXToken("CEX WBTC", "WBTC", 8, operator);

        exchange.vault().setSupportedToken(address(usdt), true);
        exchange.vault().setSupportedToken(address(wbtc), true);
        vm.stopBroadcast();

        console2.log("Exchange:", address(exchange));
        console2.log("Vault:", address(exchange.vault()));
        console2.log("Treasury:", address(exchange.treasury()));
        console2.log("USDT:", address(usdt));
        console2.log("WBTC:", address(wbtc));
        console2.log("Operator:", operator);
    }
}
