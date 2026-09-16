// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "../interfaces/IERC20.sol";

/// @title Treasury
/// @notice Holds on-chain fee reserves and cold-storage sweeps from the vault.
///         Does NOT participate in order matching — fees arrive from off-chain settlement.
contract Treasury {
    address public operator;

    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);
    event FeeReceived(address indexed token, uint256 amount, address indexed from);
    event Withdrawn(address indexed to, address indexed token, uint256 amount);

    error Unauthorized();
    error TransferFailed();

    constructor(address operator_) {
        operator = operator_;
    }

    /// @notice Accept fee sweeps from the vault operator.
    function receiveFees(address token, uint256 amount) external {
        if (!IERC20(token).transferFrom(msg.sender, address(this), amount)) {
            revert TransferFailed();
        }
        emit FeeReceived(token, amount, msg.sender);
    }

    /// @notice Operator withdraws treasury funds (cold storage, ops, etc.).
    function withdraw(address to, address token, uint256 amount) external {
        if (msg.sender != operator) revert Unauthorized();
        if (!IERC20(token).transfer(to, amount)) revert TransferFailed();
        emit Withdrawn(to, token, amount);
    }

    function setOperator(address newOperator) external {
        if (msg.sender != operator) revert Unauthorized();
        emit OperatorUpdated(operator, newOperator);
        operator = newOperator;
    }
}
