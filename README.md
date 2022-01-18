# liquidity-staking-module


The purpose of this repository is to develop and release a new set of staking, distrbution and slashing modules for the Cosmos Hub that are compatible with [cosmos-skd 0.43](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.43.0)  and do not need to wait on cosmos-sdk 0.45 to included in a Gaia release and Cosmos hub Upgrade.

## Liquid Staking design

This repo represents an opinated design for adding liquid staking to the Cosmos SDK. The

There are a few core goals.

1. Staked assets should be able to converted into liquid staked assets without unbonding.
2. The smallest change set possible on the existing staking, slashing and distrbution modules.
3. Assets are expected to be minimally fungible. Assets have a denom of "cosmosvaloperxxxx[recordId]". Record ID is a pointer to a non fungible asset that recieves the rewards from the tokenized stake.
4. Governance rights are bound to the validator.


## Typical user flow.

1. Alice bonds 500 ATOM to iqlusion.io
2. Alice executes MsgTokenizeShares for the 500_000_000uatom and 500_000_000cosmosvaloper1xxxx42 in return.
3. Alice does an OTC deal with Bob for 250_000_000cosmosvaloper1xxxx42 assets. 
4. Bob exectutes MsgRedeemTokensforShares for 250_000_000cosmosvaloper1xxxx42 which now becomes a delegation of 250atom to iqlusion.
5. While the shares were tokenized, TokenizedShareRecord 42 is recieving the full 500 atom of rewards minus iqlusion's commission.
6. Once Bob redeems his tokens for shares, now TokenizedShareRecord 42 will only recieve 250 atoms worth of rewards.
7. Alice can execute MsgWithdrawTokenizeShareRecordReward pass these rewards back to Alice's account.

### Example Commands
```bash
liquidstakingd tx staking tokenize-share cosmosvaloper1qp49fdjtlsrv6jkx3gc8urp2ncg88s6mcversm 1000000stake cosmos1qp49fdjtlsrv6jkx3gc8urp2ncg88s6macdkug
liquidstakingd query distribution tokenize-share-record-rewards cosmos1qp49fdjtlsrv6jkx3gc8urp2ncg88s6macdkug
liquidstakingd tx distribution withdraw-tokenize-share-reward

```



## Hypothetical user flow with refungiblization

This flow requires an itegration with CosmWasm that is not part of this repo at this time.

1. Alice bonds 500 ATOM to iqlusion.io 
2. Alice executes MsgTokenizeShares for the 500_000_000uatom and 500_000_000cosmosvaloper1xxxx42 in return.
3. There is a staking DAO contract in cosmwasm that is will to accept tokenizedshares from iqlusion.
4. Alice excutes a multimsg sending 500_000_000cosmosvaloper1xxxx42 and MsgTransferTokenizeShareRecord to the address of the Staking DAO contract.
5. The Staking contract queries the state to see shares to atom ratio for iqlusion and the pending rewards in the share record.
6. Alice recieve 500_000_000ustatom from the staking dao contract.


### Testnet 


Please join our testnet for Release 0.2.


[genesis.json](genesis.json)


```
persistent_peers = "6aa0b269094af7e54fbfc1faa005845fe97269e4@34.124.241.85:26656,837a38ee4fdac1b0252f31eee9780ac74d685512@34.124.168.210:26656,dd0acaa93c4bae0be874cf52ae4487551756a1e0@35.240.151.206:26656"
```