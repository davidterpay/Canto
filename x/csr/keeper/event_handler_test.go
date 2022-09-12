package keeper_test

import (
	"errors"
	"strings"

	"github.com/Canto-Network/Canto/v2/contracts"
	_ "github.com/Canto-Network/Canto/v2/x/csr/keeper"
	"github.com/Canto-Network/Canto/v2/x/csr/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// if smart contract address is not a smart contract - fail
// if smart contract has already been registered - fail
// if the receiver address does not exist - fail
// if smart contract has not yet been registered and is a contract - pass
// check that csr has been set in state
func (suite *KeeperTestSuite) TestRegisterEvent() {
	type testArgs struct {
		SmartContractAddress common.Address
		Receiver             common.Address
		ID                   uint64
	}
	suite.Commit()

	var (
		smartContractAddress = tests.GenerateAddress()
		receiver             = tests.GenerateAddress()
		turnstile, _         = suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	)

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
		setup      func()
	}{
		{
			"if smart contract address is not an account in statedb - fail",
			testArgs{
				tests.GenerateAddress(),
				tests.GenerateAddress(),
				1,
			},
			false,
			func() {},
		},
		{
			"if the smart contract address is an EOA - fail",
			testArgs{
				SmartContractAddress: smartContractAddress,
				Receiver:             receiver,
				ID:                   1,
			},
			false,
			func() {
				// set smart contract address as an EVM account
				suite.app.EvmKeeper.SetAccount(suite.ctx, smartContractAddress, *statedb.NewEmptyAccount())
			},
		},
		{
			"user is attempting to register a contract that is already registered - fail",
			testArgs{
				SmartContractAddress: smartContractAddress,
				Receiver:             receiver,
				ID:                   1,
			},
			false,
			func() {
				// set the smart contract address to a CSR
				csr := types.CSR{
					Id:        1,
					Contracts: []string{smartContractAddress.Hex()},
				}
				// set the CSR to state
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
			},
		},
		{
			"the receiver address is not a valid EVM account",
			testArgs{
				SmartContractAddress: turnstile,
				Receiver:             receiver,
				ID:                   1,
			},
			false,
			func() {
				// receiver is still not a valid account
			},
		},
		{
			"if the smart contract has not been registered yet - pass",
			testArgs{
				SmartContractAddress: turnstile,
				Receiver:             receiver,
				ID:                   2,
			},
			true,
			func() {
				// set receiver to state
				suite.app.EvmKeeper.SetAccount(suite.ctx, receiver, *statedb.NewEmptyAccount())
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// setup test
			tc.setup()
			data, err := generateRegisterEventData(tc.args.SmartContractAddress, tc.args.Receiver, tc.args.ID)
			suite.Require().NoError(err)
			// process register CSREvent
			err = suite.app.CSRKeeper.RegisterEvent(suite.ctx, data)
			if tc.expectPass {
				suite.Require().NoError(err)
				// check that the CSR exists at nftId
				csr, found := suite.app.CSRKeeper.GetCSR(suite.ctx, tc.args.ID)
				suite.Require().True(found)
				// contract address registered is correct
				suite.Require().Equal(strings.Compare(tc.args.SmartContractAddress.Hex(), csr.Contracts[0]), 0)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// if smart contract address is not a smart contract - fail
// if smart contract has already been registered - fail
// if the csr appended to does not exist - fail
// if the csr and the smart contract exist - pass
func (suite *KeeperTestSuite) TestUpdateEvent() {
	type testArgs struct {
		smartContractAddress common.Address
		nftId                uint64
	}
	suite.Commit()

	var (
		smartContractAddress = tests.GenerateAddress()
		turnstile, _         = suite.app.CSRKeeper.GetTurnstile(suite.ctx)
	)

	testCases := []struct {
		name       string
		args       testArgs
		expectPass bool
		setup      func()
	}{
		{
			"if the smart contract address is not a smart contract - fails",
			testArgs{
				smartContractAddress: smartContractAddress,
				nftId:                1,
			},
			false,
			func() {
			},
		},
		{
			"if the smart contract has alredy been registered - fail",
			testArgs{
				smartContractAddress: smartContractAddress,
				nftId:                1,
			},
			false,
			func() {
				csr := types.CSR{
					Beneficiary: sdk.AccAddress(smartContractAddress.Bytes()).String(),
					Id:          1,
					Contracts:   []string{smartContractAddress.Hex()},
				}
				// set csr to state
				suite.app.CSRKeeper.SetCSR(suite.ctx, csr)
			},
		},
		{
			"if the csr appended to does not exist - fail",
			testArgs{
				smartContractAddress: turnstile,
				nftId:                2,
			},
			false,
			func() {},
		},
		{
			"if the csr appended to exists, and the contract registered exist - pass",
			testArgs{
				smartContractAddress: turnstile,
				nftId:                1,
			},
			true,
			func() {

			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// setup test
			tc.setup()
			data, err := generateUpdateEventData(tc.args.smartContractAddress, tc.args.nftId)
			suite.Require().NoError(err)
			// process event
			err = suite.app.CSRKeeper.UpdateEvent(suite.ctx, data)
			if tc.expectPass {
				suite.Require().NoError(err)
				csr, found := suite.app.CSRKeeper.GetCSR(suite.ctx, 1)
				suite.Require().True(found)
				// contract address registered is correct
				suite.Require().Equal(strings.Compare(tc.args.smartContractAddress.Hex(), csr.Contracts[1]), 0)
			} else {
				suite.Require().Error(err)
			}
		})
	}

}

func generateUpdateEventData(contract common.Address, nftId uint64) (data []byte, err error) {
	return generateEventData("Attach", contracts.TurnstileContract, contract, nftId)
}

func generateRegisterEventData(contract, receiver common.Address, nftid uint64) (data []byte, err error) {
	return generateEventData("Register", contracts.TurnstileContract, contract, receiver, nftid)
}

// generate event creates data field for arbitrary transaction
// given a set of arguments an a method name, return the abi-encoded bytes
// of the packed event data, withdrawer, receiver, Id (not indexed)
func generateEventData(name string, contract evmtypes.CompiledContract, args ...interface{}) ([]byte, error) {
	//  retrieve arguments from contract
	var event abi.Event
	event, ok := contract.ABI.Events[name]
	if !ok {
		return nil, errors.New("cannot find event")
	}
	// ok now pack arguments
	data, err := event.Inputs.Pack(args...)
	if err != nil {
		return nil, err
	}

	return data, nil
}
