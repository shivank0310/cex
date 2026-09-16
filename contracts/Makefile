.PHONY: install test build deploy

install:
	forge install foundry-rs/forge-std

test:
	forge test -vv

build:
	forge build

deploy:
	forge script script/Deploy.s.sol --rpc-url $(EVM_RPC_URL) --broadcast
