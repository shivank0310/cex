// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {CEXVault} from "./CEXVault.sol";
import {Treasury} from "../treasury/Treasury.sol";

/// @title Exchange
/// @notice Top-level on-chain entry point for the CEX custody layer.
///
/// @dev IMPORTANT: This is NOT an on-chain order book or matching engine.
///
///      Off-chain (Go services):
///        order-service      → validates orders
///        matching-engine    → matches orders in RAM
///        settlement-service → settles trades
///        ledger-service     → double-entry accounting
///
///      On-chain (this contract):
///        deposits           → CEXVault.deposit()
///        withdrawals        → CEXVault.withdraw() via operator
///        treasury           → fee reserves
///        token operations   → CEXToken mint/transfer
contract Exchange {
    Treasury public immutable treasury;
    CEXVault public immutable vault;
    address public immutable operator;

    event Deployed(address operator, address treasury, address vault);

    constructor(address operator_) {
        operator = operator_;
        treasury = new Treasury(operator_);
        vault = new CEXVault(operator_, address(treasury));
        emit Deployed(operator_, address(treasury), address(vault));
    }
}
