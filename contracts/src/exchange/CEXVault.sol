// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "../interfaces/IERC20.sol";

/// @title CEXVault
/// @notice On-chain custody for user deposits and operator-processed withdrawals.
///
/// @dev This contract does NOT match orders. The flow is:
///      1. User deposits tokens here (on-chain)
///      2. blockchain-service monitors Deposit events
///      3. wallet-service credits the off-chain ledger
///      4. User trades via off-chain matching-engine + ledger
///      5. User requests withdrawal (off-chain ledger reservation)
///      6. Operator broadcasts withdrawal from this vault
contract CEXVault {
    address public operator;
    address public treasury;

    mapping(address => bool) public supportedTokens;
    mapping(bytes32 => bool) public processedWithdrawals;

    event OperatorUpdated(address indexed previousOperator, address indexed newOperator);
    event TreasuryUpdated(address indexed previousTreasury, address indexed newTreasury);
    event TokenSupported(address indexed token, bool supported);
    event Deposit(address indexed user, address indexed token, uint256 amount, uint256 timestamp);
    event Withdrawal(
        address indexed to,
        address indexed token,
        uint256 amount,
        bytes32 indexed withdrawalId
    );
    event FeesSwept(address indexed token, uint256 amount, address indexed treasury);

    error Unauthorized();
    error TokenNotSupported();
    error WithdrawalAlreadyProcessed();
    error TransferFailed();

    constructor(address operator_, address treasury_) {
        operator = operator_;
        treasury = treasury_;
    }

    /// @notice User deposits ERC-20 tokens into exchange custody.
    /// @dev blockchain-service watches Deposit events to credit off-chain balances.
    function deposit(address token, uint256 amount) external {
        if (!supportedTokens[token]) revert TokenNotSupported();
        if (!IERC20(token).transferFrom(msg.sender, address(this), amount)) {
            revert TransferFailed();
        }
        emit Deposit(msg.sender, token, amount, block.timestamp);
    }

    /// @notice Operator processes a user withdrawal after off-chain ledger reservation.
    /// @param withdrawalId Unique ID from wallet-service to prevent replay.
    function withdraw(address to, address token, uint256 amount, bytes32 withdrawalId) external {
        if (msg.sender != operator) revert Unauthorized();
        if (processedWithdrawals[withdrawalId]) revert WithdrawalAlreadyProcessed();
        processedWithdrawals[withdrawalId] = true;
        if (!IERC20(token).transfer(to, amount)) revert TransferFailed();
        emit Withdrawal(to, token, amount, withdrawalId);
    }

    /// @notice Sweep accumulated on-chain fees to treasury.
    function sweepFees(address token, uint256 amount) external {
        if (msg.sender != operator) revert Unauthorized();
        if (!IERC20(token).transfer(treasury, amount)) revert TransferFailed();
        emit FeesSwept(token, amount, treasury);
    }

    function setSupportedToken(address token, bool supported) external {
        if (msg.sender != operator) revert Unauthorized();
        supportedTokens[token] = supported;
        emit TokenSupported(token, supported);
    }

    function setOperator(address newOperator) external {
        if (msg.sender != operator) revert Unauthorized();
        emit OperatorUpdated(operator, newOperator);
        operator = newOperator;
    }

    function setTreasury(address newTreasury) external {
        if (msg.sender != operator) revert Unauthorized();
        emit TreasuryUpdated(treasury, newTreasury);
        treasury = newTreasury;
    }
}
