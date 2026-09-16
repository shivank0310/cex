// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {IERC20} from "../interfaces/IERC20.sol";

/// @title CEXToken
/// @notice Standard ERC-20 asset token for on-chain deposits and withdrawals.
///         Used for testnet assets (e.g. mock USDT). Trading happens off-chain.
contract CEXToken is IERC20 {
    string public name;
    string public symbol;
    uint8 public immutable decimals;

    uint256 private _totalSupply;
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    address public minter;

    event MinterUpdated(address indexed previousMinter, address indexed newMinter);

    error Unauthorized();
    error InsufficientBalance();
    error InsufficientAllowance();

    constructor(string memory name_, string memory symbol_, uint8 decimals_, address minter_) {
        name = name_;
        symbol = symbol_;
        decimals = decimals_;
        minter = minter_;
    }

    function totalSupply() external view returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) external view returns (uint256) {
        return _balances[account];
    }

    function allowance(address owner, address spender) external view returns (uint256) {
        return _allowances[owner][spender];
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        _transfer(msg.sender, to, amount);
        return true;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        _allowances[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        uint256 current = _allowances[from][msg.sender];
        if (current < amount) revert InsufficientAllowance();
        unchecked {
            _allowances[from][msg.sender] = current - amount;
        }
        _transfer(from, to, amount);
        return true;
    }

    /// @notice Mint tokens for testnet bootstrap. Production assets use bridged tokens.
    function mint(address to, uint256 amount) external {
        if (msg.sender != minter) revert Unauthorized();
        _totalSupply += amount;
        _balances[to] += amount;
        emit Transfer(address(0), to, amount);
    }

    function setMinter(address newMinter) external {
        if (msg.sender != minter) revert Unauthorized();
        emit MinterUpdated(minter, newMinter);
        minter = newMinter;
    }

    function _transfer(address from, address to, uint256 amount) internal {
        if (_balances[from] < amount) revert InsufficientBalance();
        unchecked {
            _balances[from] -= amount;
            _balances[to] += amount;
        }
        emit Transfer(from, to, amount);
    }
}
