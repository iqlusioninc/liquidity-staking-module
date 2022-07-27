package types

// staking module event types
const (
	EventTypeCompleteUnbonding           = "complete_unbonding"
	EventTypeCompleteRedelegation        = "complete_redelegation"
	EventTypeCreateValidator             = "create_validator"
	EventTypeEditValidator               = "edit_validator"
	EventTypeDelegate                    = "delegate"
	EventTypeUnbond                      = "unbond"
	EventTypeRedelegate                  = "redelegate"
	EventTypeTokenizeShares              = "tokenize_shares"
	EventTypeRedeemShares                = "redeem_shares"
	EventTypeTransferTokenizeShareRecord = "transfer_tokenize_share_record"

	AttributeKeyValidator         = "validator"
	AttributeKeyCommissionRate    = "commission_rate"
	AttributeKeyMinSelfDelegation = "min_self_delegation"
	AttributeKeySrcValidator      = "source_validator"
	AttributeKeyDstValidator      = "destination_validator"
	AttributeKeyDelegator         = "delegator"
	AttributeKeyCompletionTime    = "completion_time"
	AttributeKeyNewShares         = "new_shares"
	AttributeKeyShareOwner        = "share_owner"
	AttributeKeyShareRecordId     = "share_record_id"
	AttributeKeyAmount            = "amount"
	AttributeValueCategory        = ModuleName
)
