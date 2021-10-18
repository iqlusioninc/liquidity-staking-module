package keeper

import (
	"context"
	"fmt"
	"strconv"

	tmstrings "github.com/tendermint/tendermint/libs/strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/staking/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateValidator defines a method for creating a new validator
func (k msgServer) CreateValidator(goCtx context.Context, msg *types.MsgCreateValidator) (*types.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, valAddr); found {
		return nil, types.ErrValidatorOwnerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk)); found {
		return nil, types.ErrValidatorPubKeyExists
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Value.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Value.Denom, bondDenom,
		)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	cp := ctx.ConsensusParams()
	if cp != nil && cp.Validator != nil {
		if !tmstrings.StringInSlice(pk.Type(), cp.Validator.PubKeyTypes) {
			return nil, sdkerrors.Wrapf(
				types.ErrValidatorPubKeyTypeNotSupported,
				"got: %s, expected: %s", pk.Type(), cp.Validator.PubKeyTypes,
			)
		}
	}

	validator, err := types.NewValidator(valAddr, pk, msg.Description)
	if err != nil {
		return nil, err
	}
	commission := types.NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	if err := k.AfterValidatorCreated(ctx, validator.GetOperator()); err != nil {
		return nil, err
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return &types.MsgCreateValidatorResponse{}, err
	}

	coins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), msg.Value.Amount))
	if err := k.bankKeeper.DelegateCoinsFromAccountToModule(ctx, delegatorAddress, types.EpochDelegationPoolName, coins); err != nil {
		return &types.MsgCreateValidatorResponse{}, err
	}

	epochNumber := k.GetEpochNumber(ctx)
	k.QueueMsgForEpoch(ctx, epochNumber, msg)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
		),
	})

	return &types.MsgCreateValidatorResponse{}, nil
}

// EditValidator defines a method for editing an existing validator
func (k msgServer) EditValidator(goCtx context.Context, msg *types.MsgEditValidator) (*types.MsgEditValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Queue epoch action and move all the execution logic to Epoch execution
	epochNumber := k.GetEpochNumber(ctx)
	k.QueueMsgForEpoch(ctx, epochNumber, msg)

	cacheCtx, _ := ctx.CacheContext()
	cacheCtx = cacheCtx.WithBlockHeight(k.GetNextEpochHeight(ctx))
	cacheCtx = cacheCtx.WithBlockTime(k.GetNextEpochTime(ctx))
	err := k.executeQueuedEditValidatorMsg(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgEditValidatorResponse{}, nil
}

// Delegate defines a method for performing a delegation of coins from a delegator to a validator
func (k msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return &types.MsgDelegateResponse{}, err
	}

	bondDenom := k.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidRequest, "invalid coin denomination: got %s, expected %s", msg.Amount.Denom, bondDenom,
		)
	}

	coins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), msg.Amount.Amount))
	if err := k.bankKeeper.DelegateCoinsFromAccountToModule(ctx, delegatorAddress, types.EpochDelegationPoolName, coins); err != nil {
		return &types.MsgDelegateResponse{}, err
	}

	// Queue epoch action and move all the execution logic to Epoch execution
	epochNumber := k.GetEpochNumber(ctx)
	k.QueueMsgForEpoch(ctx, epochNumber, msg)

	// TODO should do validation by running with cachedCtx like gov proposal creation
	// To consider: cachedCtx could have status which contains all the other epoch actions
	// could add CancelDelegate since they can't do any action until Delegation finish
	return &types.MsgDelegateResponse{}, nil
}

// BeginRedelegate defines a method for performing a redelegation of coins from a delegator and source validator to a destination validator
func (k msgServer) BeginRedelegate(goCtx context.Context, msg *types.MsgBeginRedelegate) (*types.MsgBeginRedelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	epochNumber := k.GetEpochNumber(ctx)
	k.QueueMsgForEpoch(ctx, epochNumber, msg)

	cacheCtx, _ := ctx.CacheContext()
	cacheCtx = cacheCtx.WithBlockHeight(k.GetNextEpochHeight(ctx))
	cacheCtx = cacheCtx.WithBlockTime(k.GetNextEpochTime(ctx))
	completionTime, err := k.executeQueuedBeginRedelegateMsg(cacheCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgBeginRedelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

// Undelegate defines a method for performing an undelegation from a delegate and a validator
func (k msgServer) Undelegate(goCtx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.QueueMsgForEpoch(ctx, 0, msg)

	cacheCtx, _ := ctx.CacheContext()
	cacheCtx = cacheCtx.WithBlockHeight(k.GetNextEpochHeight(ctx))
	cacheCtx = cacheCtx.WithBlockTime(k.GetNextEpochTime(ctx))
	completionTime, err := k.executeQueuedUndelegateMsg(cacheCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgUndelegateResponse{
		CompletionTime: completionTime,
	}, nil
}

func getShareTokenDenom(validatorAddress string, tokenizeShareRecordId uint64) string {
	return validatorAddress + strconv.Itoa(int(tokenizeShareRecordId))
}

func (k msgServer) TokenizeShares(goCtx context.Context, msg *types.MsgTokenizeShares) (*types.MsgTokenizeSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if valErr != nil {
		return nil, valErr
	}
	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	_ = validator

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	delegation, found := k.GetDelegation(ctx, delegatorAddress, valAddr)
	if !found {
		return nil, types.ErrNoDelegatorForAddress
	}

	delegationAmount := validator.Tokens.ToDec().Mul(delegation.GetShares()).Quo(validator.DelegatorShares)
	if msg.Amount.Amount.GT(sdk.Int(delegationAmount)) {
		return nil, types.ErrNotEnoughDelegationShares
	}

	recordId := k.GetLastTokenizeShareRecordId(ctx) + 1
	shareTokenDenom := getShareTokenDenom(msg.ValidatorAddress, recordId)
	record := types.TokenizeShareRecord{
		Id:              recordId,
		Owner:           msg.DelegatorAddress,
		ShareTokenDenom: shareTokenDenom,
		ModuleAccount:   fmt.Sprintf("tokenizeshare_%d", recordId),
		Validator:       msg.ValidatorAddress,
	}

	shareToken := sdk.NewCoin(shareTokenDenom, msg.Amount.Amount)

	err = k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddress, sdk.Coins{shareToken})
	if err != nil {
		return nil, err
	}

	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valAddr, msg.Amount.Amount,
	)

	_, err = k.Unbond(ctx, delegatorAddress, valAddr, shares)
	if err != nil {
		return nil, err
	}

	// create module account
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, record.ModuleAccount)
	k.authKeeper.SetModuleAccount(ctx, moduleAcc)

	// create reward ownership record
	k.AddTokenizeShareRecord(ctx, record)

	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, moduleAcc.GetName(), sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}

	_, err = k.Keeper.Delegate(ctx, moduleAcc.GetAddress(), msg.Amount.Amount, sdkstaking.Unbonded, validator, false)
	if err != nil {
		return nil, err
	}

	return &types.MsgTokenizeSharesResponse{
		Amount: shareToken,
	}, nil
}

func (k msgServer) RedeemTokens(goCtx context.Context, msg *types.MsgRedeemTokensforShares) (*types.MsgRedeemTokensforSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	balance := k.bankKeeper.GetBalance(ctx, delegatorAddress, msg.Amount.Denom)
	if balance.Amount.LT(msg.Amount.Amount) {
		return nil, types.ErrNotEnoughBalance
	}

	record, err := k.GetTokenizeShareRecordByDenom(ctx, msg.Amount.Denom)
	if err != nil {
		return nil, err
	}

	valAddr, valErr := sdk.ValAddressFromBech32(record.Validator)
	if valErr != nil {
		return nil, valErr
	}

	validator, found := k.GetValidator(ctx, valAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	shares, err := k.ValidateUnbondAmount(
		ctx, delegatorAddress, valAddr, msg.Amount.Amount,
	)

	shareTokenModuleAccount := k.authKeeper.GetModuleAccount(ctx, record.ModuleAccount)
	_, err = k.Unbond(ctx, shareTokenModuleAccount.GetAddress(), valAddr, shares)
	if err != nil {
		return nil, err
	}

	_ = validator

	// if delegation is fully undelegated from module account, remove tokenize share record
	_, found = k.GetDelegation(ctx, shareTokenModuleAccount.GetAddress(), valAddr)
	if !found {
		k.authKeeper.RemoveAccount(ctx, shareTokenModuleAccount)
		k.DeleteTokenizeShareRecord(ctx, record.Id)
	}

	// send share tokens to NotBondedPool and burn
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, delegatorAddress, types.NotBondedPoolName, sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}
	err = k.bankKeeper.BurnCoins(ctx, types.NotBondedPoolName, sdk.Coins{msg.Amount})
	if err != nil {
		return nil, err
	}

	// send redeemed tokens to delegator address and get delegated tokens
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.NotBondedPoolName, delegatorAddress, sdk.Coins{sdk.NewCoin(k.BondDenom(ctx), msg.Amount.Amount)})
	if err != nil {
		return nil, err
	}
	_, err = k.Keeper.Delegate(ctx, delegatorAddress, msg.Amount.Amount, sdkstaking.Bonded, validator, false)
	if err != nil {
		return nil, err
	}

	return &types.MsgRedeemTokensforSharesResponse{}, nil
}

func (k msgServer) TransferTokenizeShareRecord(goCtx context.Context, msg *types.MsgTransferTokenizeShareRecord) (*types.MsgTransferTokenizeShareRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	record, err := k.GetTokenizeShareRecord(ctx, msg.TokenizeShareRecordId)
	if err != nil {
		return nil, types.ErrTokenizeShareRecordNotExists
	}

	if record.Owner != msg.Sender {
		return nil, types.ErrNotTokenizeShareRecordOwner
	}

	record.Owner = msg.NewOwner
	k.setTokenizeShareRecord(ctx, record)
	return &types.MsgTransferTokenizeShareRecordResponse{}, nil
}
