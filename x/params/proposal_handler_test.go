package params_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	govtypes "github.com/iqlusioninc/liquidity-staking-module/x/gov/types"
	"github.com/iqlusioninc/liquidity-staking-module/x/params"
	"github.com/iqlusioninc/liquidity-staking-module/x/params/types/proposal"
)

type HandlerTestSuite struct {
	suite.Suite

	app        *simapp.SimApp
	ctx        sdk.Context
	govHandler govtypes.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.app = simapp.Setup(suite.T(), false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
	suite.govHandler = params.NewParamChangeProposalHandler(suite.app.ParamsKeeper)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func testProposal(changes ...proposal.ParamChange) *proposal.ParameterChangeProposal {
	return proposal.NewParameterChangeProposal("title", "description", changes)
}

func (suite *HandlerTestSuite) TestProposalHandler() {
	testCases := []struct {
		name     string
		proposal *proposal.ParameterChangeProposal
		onHandle func()
		expErr   bool
	}{
		{
			"all fields",
			testProposal(proposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), "1")),
			func() {
				maxVals := suite.app.StakingKeeper.MaxValidators(suite.ctx)
				suite.Require().Equal(uint32(1), maxVals)
			},
			false,
		},
		{
			"invalid type",
			testProposal(proposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), "-")),
			func() {},
			true,
		},
		{
			"omit empty fields",
			testProposal(proposal.ParamChange{
				Subspace: govtypes.ModuleName,
				Key:      string(govtypes.ParamStoreKeyDepositParams),
				Value:    `{"min_deposit": [{"denom": "uatom","amount": "64000000"}]}`,
			}),
			func() {
				depositParams := suite.app.GovKeeper.GetDepositParams(suite.ctx)
				suite.Require().Equal(govtypes.DepositParams{
					MinDeposit:       sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(64000000))),
					MaxDepositPeriod: govtypes.DefaultPeriod,
				}, depositParams)
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			err := suite.govHandler(suite.ctx, tc.proposal)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				tc.onHandle()
			}
		})
	}
}