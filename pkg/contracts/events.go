package contracts

// On-chain event names emitted by contracts/src/exchange/CEXVault.sol.
// blockchain-service monitors these — NOT order or trade events.
const (
	EventVaultDeposit    = "Deposit"
	EventVaultWithdrawal = "Withdrawal"
	EventVaultFeesSwept  = "FeesSwept"
)

// VaultDepositSignature is the canonical Deposit event for log filtering.
// event Deposit(address indexed user, address indexed token, uint256 amount, uint256 timestamp);
const VaultDepositSignature = "Deposit(address,address,uint256,uint256)"
