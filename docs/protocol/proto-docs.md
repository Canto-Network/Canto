<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [canto/epochs/v1/genesis.proto](#canto/epochs/v1/genesis.proto)
    - [EpochInfo](#canto.epochs.v1.EpochInfo)
    - [GenesisState](#canto.epochs.v1.GenesisState)
  
- [canto/epochs/v1/query.proto](#canto/epochs/v1/query.proto)
    - [QueryCurrentEpochRequest](#canto.epochs.v1.QueryCurrentEpochRequest)
    - [QueryCurrentEpochResponse](#canto.epochs.v1.QueryCurrentEpochResponse)
    - [QueryEpochsInfoRequest](#canto.epochs.v1.QueryEpochsInfoRequest)
    - [QueryEpochsInfoResponse](#canto.epochs.v1.QueryEpochsInfoResponse)
  
    - [Query](#canto.epochs.v1.Query)
  
- [canto/erc20/v1/erc20.proto](#canto/erc20/v1/erc20.proto)
    - [RegisterCoinProposal](#canto.erc20.v1.RegisterCoinProposal)
    - [RegisterERC20Proposal](#canto.erc20.v1.RegisterERC20Proposal)
    - [ToggleTokenConversionProposal](#canto.erc20.v1.ToggleTokenConversionProposal)
    - [TokenPair](#canto.erc20.v1.TokenPair)
  
    - [Owner](#canto.erc20.v1.Owner)
  
- [canto/erc20/v1/genesis.proto](#canto/erc20/v1/genesis.proto)
    - [GenesisState](#canto.erc20.v1.GenesisState)
    - [Params](#canto.erc20.v1.Params)
  
- [canto/erc20/v1/query.proto](#canto/erc20/v1/query.proto)
    - [QueryParamsRequest](#canto.erc20.v1.QueryParamsRequest)
    - [QueryParamsResponse](#canto.erc20.v1.QueryParamsResponse)
    - [QueryTokenPairRequest](#canto.erc20.v1.QueryTokenPairRequest)
    - [QueryTokenPairResponse](#canto.erc20.v1.QueryTokenPairResponse)
    - [QueryTokenPairsRequest](#canto.erc20.v1.QueryTokenPairsRequest)
    - [QueryTokenPairsResponse](#canto.erc20.v1.QueryTokenPairsResponse)
  
    - [Query](#canto.erc20.v1.Query)
  
- [canto/erc20/v1/tx.proto](#canto/erc20/v1/tx.proto)
    - [MsgConvertCoin](#canto.erc20.v1.MsgConvertCoin)
    - [MsgConvertCoinResponse](#canto.erc20.v1.MsgConvertCoinResponse)
    - [MsgConvertERC20](#canto.erc20.v1.MsgConvertERC20)
    - [MsgConvertERC20Response](#canto.erc20.v1.MsgConvertERC20Response)
  
    - [Msg](#canto.erc20.v1.Msg)
  
- [canto/fees/v1/fees.proto](#canto/fees/v1/fees.proto)
    - [Fee](#canto.fees.v1.Fee)
  
- [canto/fees/v1/genesis.proto](#canto/fees/v1/genesis.proto)
    - [GenesisState](#canto.fees.v1.GenesisState)
    - [Params](#canto.fees.v1.Params)
  
- [canto/fees/v1/query.proto](#canto/fees/v1/query.proto)
    - [QueryDeployerFeesRequest](#canto.fees.v1.QueryDeployerFeesRequest)
    - [QueryDeployerFeesResponse](#canto.fees.v1.QueryDeployerFeesResponse)
    - [QueryFeeRequest](#canto.fees.v1.QueryFeeRequest)
    - [QueryFeeResponse](#canto.fees.v1.QueryFeeResponse)
    - [QueryFeesRequest](#canto.fees.v1.QueryFeesRequest)
    - [QueryFeesResponse](#canto.fees.v1.QueryFeesResponse)
    - [QueryParamsRequest](#canto.fees.v1.QueryParamsRequest)
    - [QueryParamsResponse](#canto.fees.v1.QueryParamsResponse)
  
    - [Query](#canto.fees.v1.Query)
  
- [canto/fees/v1/tx.proto](#canto/fees/v1/tx.proto)
    - [MsgCancelFee](#canto.fees.v1.MsgCancelFee)
    - [MsgCancelFeeResponse](#canto.fees.v1.MsgCancelFeeResponse)
    - [MsgRegisterFee](#canto.fees.v1.MsgRegisterFee)
    - [MsgRegisterFeeResponse](#canto.fees.v1.MsgRegisterFeeResponse)
    - [MsgUpdateFee](#canto.fees.v1.MsgUpdateFee)
    - [MsgUpdateFeeResponse](#canto.fees.v1.MsgUpdateFeeResponse)
  
    - [Msg](#canto.fees.v1.Msg)
  
- [canto/govshuttle/v1/govshuttle.proto](#canto/govshuttle/v1/govshuttle.proto)
    - [LendingMarketMetadata](#canto.govshuttle.v1.LendingMarketMetadata)
    - [LendingMarketProposal](#canto.govshuttle.v1.LendingMarketProposal)
    - [Params](#canto.govshuttle.v1.Params)
    - [TreasuryProposal](#canto.govshuttle.v1.TreasuryProposal)
    - [TreasuryProposalMetadata](#canto.govshuttle.v1.TreasuryProposalMetadata)
  
- [canto/govshuttle/v1/genesis.proto](#canto/govshuttle/v1/genesis.proto)
    - [GenesisState](#canto.govshuttle.v1.GenesisState)
  
- [canto/govshuttle/v1/query.proto](#canto/govshuttle/v1/query.proto)
    - [QueryParamsRequest](#canto.govshuttle.v1.QueryParamsRequest)
    - [QueryParamsResponse](#canto.govshuttle.v1.QueryParamsResponse)
  
    - [Query](#canto.govshuttle.v1.Query)
  
- [canto/govshuttle/v1/tx.proto](#canto/govshuttle/v1/tx.proto)
    - [Msg](#canto.govshuttle.v1.Msg)
  
- [canto/inflation/v1/inflation.proto](#canto/inflation/v1/inflation.proto)
    - [ExponentialCalculation](#canto.inflation.v1.ExponentialCalculation)
    - [InflationDistribution](#canto.inflation.v1.InflationDistribution)
  
- [canto/inflation/v1/genesis.proto](#canto/inflation/v1/genesis.proto)
    - [GenesisState](#canto.inflation.v1.GenesisState)
    - [Params](#canto.inflation.v1.Params)
  
- [canto/inflation/v1/query.proto](#canto/inflation/v1/query.proto)
    - [QueryCirculatingSupplyRequest](#canto.inflation.v1.QueryCirculatingSupplyRequest)
    - [QueryCirculatingSupplyResponse](#canto.inflation.v1.QueryCirculatingSupplyResponse)
    - [QueryEpochMintProvisionRequest](#canto.inflation.v1.QueryEpochMintProvisionRequest)
    - [QueryEpochMintProvisionResponse](#canto.inflation.v1.QueryEpochMintProvisionResponse)
    - [QueryInflationRateRequest](#canto.inflation.v1.QueryInflationRateRequest)
    - [QueryInflationRateResponse](#canto.inflation.v1.QueryInflationRateResponse)
    - [QueryParamsRequest](#canto.inflation.v1.QueryParamsRequest)
    - [QueryParamsResponse](#canto.inflation.v1.QueryParamsResponse)
    - [QueryPeriodRequest](#canto.inflation.v1.QueryPeriodRequest)
    - [QueryPeriodResponse](#canto.inflation.v1.QueryPeriodResponse)
    - [QuerySkippedEpochsRequest](#canto.inflation.v1.QuerySkippedEpochsRequest)
    - [QuerySkippedEpochsResponse](#canto.inflation.v1.QuerySkippedEpochsResponse)
  
    - [Query](#canto.inflation.v1.Query)
  
- [canto/recovery/v1/genesis.proto](#canto/recovery/v1/genesis.proto)
    - [GenesisState](#canto.recovery.v1.GenesisState)
    - [Params](#canto.recovery.v1.Params)
  
- [canto/recovery/v1/query.proto](#canto/recovery/v1/query.proto)
    - [QueryParamsRequest](#canto.recovery.v1.QueryParamsRequest)
    - [QueryParamsResponse](#canto.recovery.v1.QueryParamsResponse)
  
    - [Query](#canto.recovery.v1.Query)
  
- [canto/vesting/v1/query.proto](#canto/vesting/v1/query.proto)
    - [QueryBalancesRequest](#canto.vesting.v1.QueryBalancesRequest)
    - [QueryBalancesResponse](#canto.vesting.v1.QueryBalancesResponse)
  
    - [Query](#canto.vesting.v1.Query)
  
- [canto/vesting/v1/tx.proto](#canto/vesting/v1/tx.proto)
    - [MsgClawback](#canto.vesting.v1.MsgClawback)
    - [MsgClawbackResponse](#canto.vesting.v1.MsgClawbackResponse)
    - [MsgCreateClawbackVestingAccount](#canto.vesting.v1.MsgCreateClawbackVestingAccount)
    - [MsgCreateClawbackVestingAccountResponse](#canto.vesting.v1.MsgCreateClawbackVestingAccountResponse)
  
    - [Msg](#canto.vesting.v1.Msg)
  
- [canto/vesting/v1/vesting.proto](#canto/vesting/v1/vesting.proto)
    - [ClawbackVestingAccount](#canto.vesting.v1.ClawbackVestingAccount)
  
- [Scalar Value Types](#scalar-value-types)



<a name="canto/epochs/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/epochs/v1/genesis.proto



<a name="canto.epochs.v1.EpochInfo"></a>

### EpochInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifier` | [string](#string) |  |  |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `current_epoch` | [int64](#int64) |  |  |
| `current_epoch_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `epoch_counting_started` | [bool](#bool) |  |  |
| `current_epoch_start_height` | [int64](#int64) |  |  |






<a name="canto.epochs.v1.GenesisState"></a>

### GenesisState
GenesisState defines the epochs module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epochs` | [EpochInfo](#canto.epochs.v1.EpochInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/epochs/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/epochs/v1/query.proto



<a name="canto.epochs.v1.QueryCurrentEpochRequest"></a>

### QueryCurrentEpochRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifier` | [string](#string) |  |  |






<a name="canto.epochs.v1.QueryCurrentEpochResponse"></a>

### QueryCurrentEpochResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `current_epoch` | [int64](#int64) |  |  |






<a name="canto.epochs.v1.QueryEpochsInfoRequest"></a>

### QueryEpochsInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="canto.epochs.v1.QueryEpochsInfoResponse"></a>

### QueryEpochsInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epochs` | [EpochInfo](#canto.epochs.v1.EpochInfo) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.epochs.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `EpochInfos` | [QueryEpochsInfoRequest](#canto.epochs.v1.QueryEpochsInfoRequest) | [QueryEpochsInfoResponse](#canto.epochs.v1.QueryEpochsInfoResponse) | EpochInfos provide running epochInfos | GET|/canto/epochs/v1/epochs|
| `CurrentEpoch` | [QueryCurrentEpochRequest](#canto.epochs.v1.QueryCurrentEpochRequest) | [QueryCurrentEpochResponse](#canto.epochs.v1.QueryCurrentEpochResponse) | CurrentEpoch provide current epoch of specified identifier | GET|/canto/epochs/v1/current_epoch|

 <!-- end services -->



<a name="canto/erc20/v1/erc20.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/erc20/v1/erc20.proto



<a name="canto.erc20.v1.RegisterCoinProposal"></a>

### RegisterCoinProposal
RegisterCoinProposal is a gov Content type to register a token pair for a
native Cosmos coin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos.bank.v1beta1.Metadata) |  | metadata of the native Cosmos coin |






<a name="canto.erc20.v1.RegisterERC20Proposal"></a>

### RegisterERC20Proposal
RegisterERC20Proposal is a gov Content type to register a token pair for an
ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposa string title = 1; |
| `description` | [string](#string) |  | proposal description |
| `erc20address` | [string](#string) |  | contract address of ERC20 token |






<a name="canto.erc20.v1.ToggleTokenConversionProposal"></a>

### ToggleTokenConversionProposal
ToggleTokenConversionProposal is a gov Content type to toggle the conversion
of a token pair.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `Title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `token` | [string](#string) |  | token identifier can be either the hex contract address of the ERC20 or the Cosmos base denomination |






<a name="canto.erc20.v1.TokenPair"></a>

### TokenPair
TokenPair defines an instance that records a pairing consisting of a native
 Cosmos Coin and an ERC20 token address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `erc20_address` | [string](#string) |  | address of ERC20 contract token |
| `denom` | [string](#string) |  | cosmos base denomination to be mapped to |
| `enabled` | [bool](#bool) |  | shows token mapping enable status |
| `contract_owner` | [Owner](#canto.erc20.v1.Owner) |  | ERC20 owner address ENUM (0 invalid, 1 ModuleAccount, 2 external address) |





 <!-- end messages -->


<a name="canto.erc20.v1.Owner"></a>

### Owner
Owner enumerates the ownership of a ERC20 contract.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OWNER_UNSPECIFIED | 0 | OWNER_UNSPECIFIED defines an invalid/undefined owner. |
| OWNER_MODULE | 1 | OWNER_MODULE erc20 is owned by the erc20 module account. |
| OWNER_EXTERNAL | 2 | EXTERNAL erc20 is owned by an external account. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/erc20/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/erc20/v1/genesis.proto



<a name="canto.erc20.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.erc20.v1.Params) |  | module parameters |
| `token_pairs` | [TokenPair](#canto.erc20.v1.TokenPair) | repeated | registered token pairs |






<a name="canto.erc20.v1.Params"></a>

### Params
Params defines the erc20 module params


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_erc20` | [bool](#bool) |  | parameter to enable the conversion of Cosmos coins <--> ERC20 tokens. |
| `enable_evm_hook` | [bool](#bool) |  | parameter to enable the EVM hook that converts an ERC20 token to a Cosmos Coin by transferring the Tokens through a MsgEthereumTx to the ModuleAddress Ethereum address. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/erc20/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/erc20/v1/query.proto



<a name="canto.erc20.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="canto.erc20.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.erc20.v1.Params) |  |  |






<a name="canto.erc20.v1.QueryTokenPairRequest"></a>

### QueryTokenPairRequest
QueryTokenPairRequest is the request type for the Query/TokenPair RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token` | [string](#string) |  | token identifier can be either the hex contract address of the ERC20 or the Cosmos base denomination |






<a name="canto.erc20.v1.QueryTokenPairResponse"></a>

### QueryTokenPairResponse
QueryTokenPairResponse is the response type for the Query/TokenPair RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_pair` | [TokenPair](#canto.erc20.v1.TokenPair) |  |  |






<a name="canto.erc20.v1.QueryTokenPairsRequest"></a>

### QueryTokenPairsRequest
QueryTokenPairsRequest is the request type for the Query/TokenPairs RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="canto.erc20.v1.QueryTokenPairsResponse"></a>

### QueryTokenPairsResponse
QueryTokenPairsResponse is the response type for the Query/TokenPairs RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_pairs` | [TokenPair](#canto.erc20.v1.TokenPair) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.erc20.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `TokenPairs` | [QueryTokenPairsRequest](#canto.erc20.v1.QueryTokenPairsRequest) | [QueryTokenPairsResponse](#canto.erc20.v1.QueryTokenPairsResponse) | TokenPairs retrieves registered token pairs | GET|/canto/erc20/v1/token_pairs|
| `TokenPair` | [QueryTokenPairRequest](#canto.erc20.v1.QueryTokenPairRequest) | [QueryTokenPairResponse](#canto.erc20.v1.QueryTokenPairResponse) | TokenPair retrieves a registered token pair | GET|/canto/erc20/v1/token_pairs/{token}|
| `Params` | [QueryParamsRequest](#canto.erc20.v1.QueryParamsRequest) | [QueryParamsResponse](#canto.erc20.v1.QueryParamsResponse) | Params retrieves the erc20 module params | GET|/canto/erc20/v1/params|

 <!-- end services -->



<a name="canto/erc20/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/erc20/v1/tx.proto



<a name="canto.erc20.v1.MsgConvertCoin"></a>

### MsgConvertCoin
MsgConvertCoin defines a Msg to convert a native Cosmos coin to a ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Cosmos coin which denomination is registered in a token pair. The coin amount defines the amount of coins to convert. |
| `receiver` | [string](#string) |  | recipient hex address to receive ERC20 token |
| `sender` | [string](#string) |  | cosmos bech32 address from the owner of the given Cosmos coins |






<a name="canto.erc20.v1.MsgConvertCoinResponse"></a>

### MsgConvertCoinResponse
MsgConvertCoinResponse returns no fields






<a name="canto.erc20.v1.MsgConvertERC20"></a>

### MsgConvertERC20
MsgConvertERC20 defines a Msg to convert a ERC20 token to a native Cosmos
coin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | ERC20 token contract address registered in a token pair |
| `amount` | [string](#string) |  | amount of ERC20 tokens to convert |
| `receiver` | [string](#string) |  | bech32 address to receive native Cosmos coins |
| `sender` | [string](#string) |  | sender hex address from the owner of the given ERC20 tokens |






<a name="canto.erc20.v1.MsgConvertERC20Response"></a>

### MsgConvertERC20Response
MsgConvertERC20Response returns no fields





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.erc20.v1.Msg"></a>

### Msg
Msg defines the erc20 Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ConvertCoin` | [MsgConvertCoin](#canto.erc20.v1.MsgConvertCoin) | [MsgConvertCoinResponse](#canto.erc20.v1.MsgConvertCoinResponse) | ConvertCoin mints a ERC20 representation of the native Cosmos coin denom that is registered on the token mapping. | GET|/canto/erc20/v1/tx/convert_coin|
| `ConvertERC20` | [MsgConvertERC20](#canto.erc20.v1.MsgConvertERC20) | [MsgConvertERC20Response](#canto.erc20.v1.MsgConvertERC20Response) | ConvertERC20 mints a native Cosmos coin representation of the ERC20 token contract that is registered on the token mapping. | GET|/canto/erc20/v1/tx/convert_erc20|

 <!-- end services -->



<a name="canto/fees/v1/fees.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/fees/v1/fees.proto



<a name="canto.fees.v1.Fee"></a>

### Fee
Fee defines an instance that organizes fee distribution conditions for the
owner of a given smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | hex address of registered contract |
| `deployer_address` | [string](#string) |  | bech32 address of contract deployer |
| `withdraw_address` | [string](#string) |  | bech32 address of account receiving the transaction fees it defaults to deployer_address |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/fees/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/fees/v1/genesis.proto



<a name="canto.fees.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.fees.v1.Params) |  | module parameters |
| `fees` | [Fee](#canto.fees.v1.Fee) | repeated | active registered contracts for fee distribution |






<a name="canto.fees.v1.Params"></a>

### Params
Params defines the fees module params


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_fees` | [bool](#bool) |  | parameter to enable fees |
| `developer_shares` | [string](#string) |  | developer_shares defines the proportion of the transaction fees to be distributed to the registered contract owner |
| `addr_derivation_cost_create` | [uint64](#uint64) |  | addr_derivation_cost_create defines the cost of address derivation for verifying the contract deployer at fee registration |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/fees/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/fees/v1/query.proto



<a name="canto.fees.v1.QueryDeployerFeesRequest"></a>

### QueryDeployerFeesRequest
QueryDeployerFeesRequest is the request type for the Query/DeployerFees RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `deployer_address` | [string](#string) |  | deployer bech32 address |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="canto.fees.v1.QueryDeployerFeesResponse"></a>

### QueryDeployerFeesResponse
QueryDeployerFeesResponse is the response type for the Query/DeployerFees RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fees` | [Fee](#canto.fees.v1.Fee) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="canto.fees.v1.QueryFeeRequest"></a>

### QueryFeeRequest
QueryFeeRequest is the request type for the Query/Fee RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract identifier is the hex contract address of a contract |






<a name="canto.fees.v1.QueryFeeResponse"></a>

### QueryFeeResponse
QueryFeeResponse is the response type for the Query/Fee RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fee` | [Fee](#canto.fees.v1.Fee) |  |  |






<a name="canto.fees.v1.QueryFeesRequest"></a>

### QueryFeesRequest
QueryFeesRequest is the request type for the Query/Fees RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="canto.fees.v1.QueryFeesResponse"></a>

### QueryFeesResponse
QueryFeesResponse is the response type for the Query/Fees RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `fees` | [Fee](#canto.fees.v1.Fee) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="canto.fees.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="canto.fees.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.fees.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.fees.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Fees` | [QueryFeesRequest](#canto.fees.v1.QueryFeesRequest) | [QueryFeesResponse](#canto.fees.v1.QueryFeesResponse) | Fees retrieves all registered contracts for fee distribution | GET|/canto/fees/v1/fees|
| `Fee` | [QueryFeeRequest](#canto.fees.v1.QueryFeeRequest) | [QueryFeeResponse](#canto.fees.v1.QueryFeeResponse) | Fee retrieves a registered contract for fee distribution for a given address | GET|/canto/fees/v1/fees/{contract_address}|
| `Params` | [QueryParamsRequest](#canto.fees.v1.QueryParamsRequest) | [QueryParamsResponse](#canto.fees.v1.QueryParamsResponse) | Params retrieves the fees module params | GET|/canto/fees/v1/params|
| `DeployerFees` | [QueryDeployerFeesRequest](#canto.fees.v1.QueryDeployerFeesRequest) | [QueryDeployerFeesResponse](#canto.fees.v1.QueryDeployerFeesResponse) | DeployerFees retrieves all contracts that a given deployer has registered for fee distribution | GET|/canto/fees/v1/fees/{deployer_address}|

 <!-- end services -->



<a name="canto/fees/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/fees/v1/tx.proto



<a name="canto.fees.v1.MsgCancelFee"></a>

### MsgCancelFee
MsgCancelFee defines a message that cancels a registered a Fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract hex address |
| `deployer_address` | [string](#string) |  | deployer bech32 address |






<a name="canto.fees.v1.MsgCancelFeeResponse"></a>

### MsgCancelFeeResponse
MsgCancelFeeResponse defines the MsgCancelFee response type






<a name="canto.fees.v1.MsgRegisterFee"></a>

### MsgRegisterFee
MsgRegisterFee defines a message that registers a Fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract hex address |
| `deployer_address` | [string](#string) |  | bech32 address of message sender, must be the same as the origin EOA sending the transaction which deploys the contract |
| `withdraw_address` | [string](#string) |  | bech32 address of account receiving the transaction fees |
| `nonces` | [uint64](#uint64) | repeated | array of nonces from the address path, where the last nonce is the nonce that determines the contract's address - it can be an EOA nonce or a factory contract nonce |






<a name="canto.fees.v1.MsgRegisterFeeResponse"></a>

### MsgRegisterFeeResponse
MsgRegisterFeeResponse defines the MsgRegisterFee response type






<a name="canto.fees.v1.MsgUpdateFee"></a>

### MsgUpdateFee
MsgUpdateFee defines a message that updates the withdraw address for a
registered Fee


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | contract hex address |
| `deployer_address` | [string](#string) |  | deployer bech32 address |
| `withdraw_address` | [string](#string) |  | new withdraw bech32 address for receiving the transaction fees |






<a name="canto.fees.v1.MsgUpdateFeeResponse"></a>

### MsgUpdateFeeResponse
MsgUpdateFeeResponse defines the MsgUpdateFee response type





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.fees.v1.Msg"></a>

### Msg
Msg defines the fees Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `RegisterFee` | [MsgRegisterFee](#canto.fees.v1.MsgRegisterFee) | [MsgRegisterFeeResponse](#canto.fees.v1.MsgRegisterFeeResponse) | RegisterFee registers a new contract for receiving transaction fees | POST|/canto/fees/v1/tx/register_fee|
| `CancelFee` | [MsgCancelFee](#canto.fees.v1.MsgCancelFee) | [MsgCancelFeeResponse](#canto.fees.v1.MsgCancelFeeResponse) | CancelFee cancels a contract's fee registration and further receival of transaction fees | POST|/canto/fees/v1/tx/cancel_fee|
| `UpdateFee` | [MsgUpdateFee](#canto.fees.v1.MsgUpdateFee) | [MsgUpdateFeeResponse](#canto.fees.v1.MsgUpdateFeeResponse) | UpdateFee updates the withdraw address | POST|/canto/fees/v1/tx/update_fee|

 <!-- end services -->



<a name="canto/govshuttle/v1/govshuttle.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/govshuttle/v1/govshuttle.proto



<a name="canto.govshuttle.v1.LendingMarketMetadata"></a>

### LendingMarketMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `Account` | [string](#string) | repeated |  |
| `PropId` | [uint64](#uint64) |  |  |
| `values` | [uint64](#uint64) | repeated |  |
| `calldatas` | [string](#string) | repeated |  |
| `signatures` | [string](#string) | repeated |  |






<a name="canto.govshuttle.v1.LendingMarketProposal"></a>

### LendingMarketProposal
Define this object so that the govshuttle.pb.go file is generate, implements
govtypes.Content


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title |
| `description` | [string](#string) |  |  |
| `metadata` | [LendingMarketMetadata](#canto.govshuttle.v1.LendingMarketMetadata) |  |  |






<a name="canto.govshuttle.v1.Params"></a>

### Params
Params defines the parameters for the module.






<a name="canto.govshuttle.v1.TreasuryProposal"></a>

### TreasuryProposal
treasury proposal type,


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  |  |
| `description` | [string](#string) |  |  |
| `metadata` | [TreasuryProposalMetadata](#canto.govshuttle.v1.TreasuryProposalMetadata) |  |  |






<a name="canto.govshuttle.v1.TreasuryProposalMetadata"></a>

### TreasuryProposalMetadata



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `PropID` | [uint64](#uint64) |  | proposalID, for querying proposals in EVM side, |
| `recipient` | [string](#string) |  | bytestring representing account addresses |
| `amount` | [uint64](#uint64) |  |  |
| `denom` | [string](#string) |  | canto or note |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/govshuttle/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/govshuttle/v1/genesis.proto



<a name="canto.govshuttle.v1.GenesisState"></a>

### GenesisState
GenesisState defines the govshuttle module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.govshuttle.v1.Params) |  | this line is used by starport scaffolding # genesis/proto/state |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/govshuttle/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/govshuttle/v1/query.proto



<a name="canto.govshuttle.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is request type for the Query/Params RPC method.






<a name="canto.govshuttle.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.govshuttle.v1.Params) |  | params holds all the parameters of this module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.govshuttle.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#canto.govshuttle.v1.QueryParamsRequest) | [QueryParamsResponse](#canto.govshuttle.v1.QueryParamsResponse) | Parameters queries the parameters of the module. | GET|/canto/govshuttle/params|

 <!-- end services -->



<a name="canto/govshuttle/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/govshuttle/v1/tx.proto


 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.govshuttle.v1.Msg"></a>

### Msg
Msg defines the Msg service.

this line is used by starport scaffolding # proto/tx/rpc

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |

 <!-- end services -->



<a name="canto/inflation/v1/inflation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/inflation/v1/inflation.proto



<a name="canto.inflation.v1.ExponentialCalculation"></a>

### ExponentialCalculation
ExponentialCalculation holds factors to calculate exponential inflation on
each period. Calculation reference:
periodProvision = exponentialDecay       *  bondingIncentive
f(x)            = (a * (1 - r) ^ x + c)  *  (1 + max_variance - bondedRatio *
(max_variance / bonding_target))


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `a` | [string](#string) |  | initial value |
| `r` | [string](#string) |  | reduction factor |
| `c` | [string](#string) |  | long term inflation |
| `bonding_target` | [string](#string) |  | bonding target |
| `max_variance` | [string](#string) |  | max variance |






<a name="canto.inflation.v1.InflationDistribution"></a>

### InflationDistribution
InflationDistribution defines the distribution in which inflation is
allocated through minting on each epoch (staking, incentives, community). It
excludes the team vesting distribution, as this is minted once at genesis.
The initial InflationDistribution can be calculated from the Evmos Token
Model like this:
mintDistribution1 = distribution1 / (1 - teamVestingDistribution)
0.5333333         = 40%           / (1 - 25%)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `staking_rewards` | [string](#string) |  | staking_rewards defines the proportion of the minted minted_denom that is to be allocated as staking rewards |
| `community_pool` | [string](#string) |  | usage_incentives defines the proportion of the minted minted_denom that is // to be allocated to the incentives module address string usage_incentives = 2 [ (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Dec", (gogoproto.nullable) = false ]; community_pool defines the proportion of the minted minted_denom that is to be allocated to the community pool |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/inflation/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/inflation/v1/genesis.proto



<a name="canto.inflation.v1.GenesisState"></a>

### GenesisState
GenesisState defines the inflation module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.inflation.v1.Params) |  | params defines all the paramaters of the module. |
| `period` | [uint64](#uint64) |  | amount of past periods, based on the epochs per period param |
| `epoch_identifier` | [string](#string) |  | inflation epoch identifier |
| `epochs_per_period` | [int64](#int64) |  | number of epochs after which inflation is recalculated |
| `skipped_epochs` | [uint64](#uint64) |  | number of epochs that have passed while inflation is disabled |






<a name="canto.inflation.v1.Params"></a>

### Params
Params holds parameters for the inflation module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mint_denom` | [string](#string) |  | type of coin to mint |
| `exponential_calculation` | [ExponentialCalculation](#canto.inflation.v1.ExponentialCalculation) |  | variables to calculate exponential inflation |
| `inflation_distribution` | [InflationDistribution](#canto.inflation.v1.InflationDistribution) |  | inflation distribution of the minted denom |
| `enable_inflation` | [bool](#bool) |  | parameter to enable inflation and halt increasing the skipped_epochs |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/inflation/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/inflation/v1/query.proto



<a name="canto.inflation.v1.QueryCirculatingSupplyRequest"></a>

### QueryCirculatingSupplyRequest
QueryCirculatingSupplyRequest is the request type for the
Query/CirculatingSupply RPC method.






<a name="canto.inflation.v1.QueryCirculatingSupplyResponse"></a>

### QueryCirculatingSupplyResponse
QueryCirculatingSupplyResponse is the response type for the
Query/CirculatingSupply RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `circulating_supply` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | total amount of coins in circulation |






<a name="canto.inflation.v1.QueryEpochMintProvisionRequest"></a>

### QueryEpochMintProvisionRequest
QueryEpochMintProvisionRequest is the request type for the
Query/EpochMintProvision RPC method.






<a name="canto.inflation.v1.QueryEpochMintProvisionResponse"></a>

### QueryEpochMintProvisionResponse
QueryEpochMintProvisionResponse is the response type for the
Query/EpochMintProvision RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epoch_mint_provision` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  | epoch_mint_provision is the current minting per epoch provision value. |






<a name="canto.inflation.v1.QueryInflationRateRequest"></a>

### QueryInflationRateRequest
QueryInflationRateRequest is the request type for the Query/InflationRate RPC
method.






<a name="canto.inflation.v1.QueryInflationRateResponse"></a>

### QueryInflationRateResponse
QueryInflationRateResponse is the response type for the Query/InflationRate
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `inflation_rate` | [string](#string) |  | rate by which the total supply increases within one period |






<a name="canto.inflation.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="canto.inflation.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.inflation.v1.Params) |  | params defines the parameters of the module. |






<a name="canto.inflation.v1.QueryPeriodRequest"></a>

### QueryPeriodRequest
QueryPeriodRequest is the request type for the Query/Period RPC method.






<a name="canto.inflation.v1.QueryPeriodResponse"></a>

### QueryPeriodResponse
QueryPeriodResponse is the response type for the Query/Period RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `period` | [uint64](#uint64) |  | period is the current minting per epoch provision value. |






<a name="canto.inflation.v1.QuerySkippedEpochsRequest"></a>

### QuerySkippedEpochsRequest
QuerySkippedEpochsRequest is the request type for the Query/SkippedEpochs RPC
method.






<a name="canto.inflation.v1.QuerySkippedEpochsResponse"></a>

### QuerySkippedEpochsResponse
QuerySkippedEpochsResponse is the response type for the Query/SkippedEpochs
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `skipped_epochs` | [uint64](#uint64) |  | number of epochs that the inflation module has been disabled. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.inflation.v1.Query"></a>

### Query
Query provides defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Period` | [QueryPeriodRequest](#canto.inflation.v1.QueryPeriodRequest) | [QueryPeriodResponse](#canto.inflation.v1.QueryPeriodResponse) | Period retrieves current period. | GET|/canto/inflation/v1/period|
| `EpochMintProvision` | [QueryEpochMintProvisionRequest](#canto.inflation.v1.QueryEpochMintProvisionRequest) | [QueryEpochMintProvisionResponse](#canto.inflation.v1.QueryEpochMintProvisionResponse) | EpochMintProvision retrieves current minting epoch provision value. | GET|/canto/inflation/v1/epoch_mint_provision|
| `SkippedEpochs` | [QuerySkippedEpochsRequest](#canto.inflation.v1.QuerySkippedEpochsRequest) | [QuerySkippedEpochsResponse](#canto.inflation.v1.QuerySkippedEpochsResponse) | SkippedEpochs retrieves the total number of skipped epochs. | GET|/canto/inflation/v1/skipped_epochs|
| `CirculatingSupply` | [QueryCirculatingSupplyRequest](#canto.inflation.v1.QueryCirculatingSupplyRequest) | [QueryCirculatingSupplyResponse](#canto.inflation.v1.QueryCirculatingSupplyResponse) | CirculatingSupply retrieves the total number of tokens that are in circulation (i.e. excluding unvested tokens). | GET|/canto/inflation/v1/circulating_supply|
| `InflationRate` | [QueryInflationRateRequest](#canto.inflation.v1.QueryInflationRateRequest) | [QueryInflationRateResponse](#canto.inflation.v1.QueryInflationRateResponse) | InflationRate retrieves the inflation rate of the current period. | GET|/canto/inflation/v1/inflation_rate|
| `Params` | [QueryParamsRequest](#canto.inflation.v1.QueryParamsRequest) | [QueryParamsResponse](#canto.inflation.v1.QueryParamsResponse) | Params retrieves the total set of minting parameters. | GET|/canto/inflation/v1/params|

 <!-- end services -->



<a name="canto/recovery/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/recovery/v1/genesis.proto



<a name="canto.recovery.v1.GenesisState"></a>

### GenesisState
GenesisState defines the recovery module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.recovery.v1.Params) |  | params defines all the paramaters of the module. |






<a name="canto.recovery.v1.Params"></a>

### Params
Params holds parameters for the recovery module


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_recovery` | [bool](#bool) |  | enable recovery IBC middleware |
| `packet_timeout_duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  | duration added to timeout timestamp for balances recovered via IBC packets |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="canto/recovery/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/recovery/v1/query.proto



<a name="canto.recovery.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="canto.recovery.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#canto.recovery.v1.Params) |  | params defines the parameters of the module. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.recovery.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Params` | [QueryParamsRequest](#canto.recovery.v1.QueryParamsRequest) | [QueryParamsResponse](#canto.recovery.v1.QueryParamsResponse) | Params retrieves the total set of recovery parameters. | GET|/canto/recovery/v1/params|

 <!-- end services -->



<a name="canto/vesting/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/vesting/v1/query.proto



<a name="canto.vesting.v1.QueryBalancesRequest"></a>

### QueryBalancesRequest
QueryBalancesRequest is the request type for the Query/Balances RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | address of the clawback vesting account |






<a name="canto.vesting.v1.QueryBalancesResponse"></a>

### QueryBalancesResponse
QueryBalancesResponse is the response type for the Query/Balances RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `locked` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | current amount of locked tokens |
| `unvested` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | current amount of unvested tokens |
| `vested` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | current amount of vested tokens |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.vesting.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Balances` | [QueryBalancesRequest](#canto.vesting.v1.QueryBalancesRequest) | [QueryBalancesResponse](#canto.vesting.v1.QueryBalancesResponse) | Retrieves the unvested, vested and locked tokens for a vesting account | GET|/canto/vesting/v1/balances/{address}|

 <!-- end services -->



<a name="canto/vesting/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/vesting/v1/tx.proto



<a name="canto.vesting.v1.MsgClawback"></a>

### MsgClawback
MsgClawback defines a message that removes unvested tokens from a
ClawbackVestingAccount.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `funder_address` | [string](#string) |  | funder_address is the address which funded the account |
| `account_address` | [string](#string) |  | account_address is the address of the ClawbackVestingAccount to claw back from. |
| `dest_address` | [string](#string) |  | dest_address specifies where the clawed-back tokens should be transferred to. If empty, the tokens will be transferred back to the original funder of the account. |






<a name="canto.vesting.v1.MsgClawbackResponse"></a>

### MsgClawbackResponse
MsgClawbackResponse defines the MsgClawback response type.






<a name="canto.vesting.v1.MsgCreateClawbackVestingAccount"></a>

### MsgCreateClawbackVestingAccount
MsgCreateClawbackVestingAccount defines a message that enables creating a
ClawbackVestingAccount.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `from_address` | [string](#string) |  | from_address specifies the account to provide the funds and sign the clawback request |
| `to_address` | [string](#string) |  | to_address specifies the account to receive the funds |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | start_time defines the time at which the vesting period begins |
| `lockup_periods` | [cosmos.vesting.v1beta1.Period](#cosmos.vesting.v1beta1.Period) | repeated | lockup_periods defines the unlocking schedule relative to the start_time |
| `vesting_periods` | [cosmos.vesting.v1beta1.Period](#cosmos.vesting.v1beta1.Period) | repeated | vesting_periods defines thevesting schedule relative to the start_time |
| `merge` | [bool](#bool) |  | merge specifies a the creation mechanism for existing ClawbackVestingAccounts. If true, merge this new grant into an existing ClawbackVestingAccount, or create it if it does not exist. If false, creates a new account. New grants to an existing account must be from the same from_address. |






<a name="canto.vesting.v1.MsgCreateClawbackVestingAccountResponse"></a>

### MsgCreateClawbackVestingAccountResponse
MsgCreateClawbackVestingAccountResponse defines the
MsgCreateClawbackVestingAccount response type.





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="canto.vesting.v1.Msg"></a>

### Msg
Msg defines the vesting Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `CreateClawbackVestingAccount` | [MsgCreateClawbackVestingAccount](#canto.vesting.v1.MsgCreateClawbackVestingAccount) | [MsgCreateClawbackVestingAccountResponse](#canto.vesting.v1.MsgCreateClawbackVestingAccountResponse) | CreateClawbackVestingAccount creats a vesting account that is subject to clawback and the configuration of vesting and lockup schedules. | GET|/canto/vesting/v1/tx/create_clawback_vesting_account|
| `Clawback` | [MsgClawback](#canto.vesting.v1.MsgClawback) | [MsgClawbackResponse](#canto.vesting.v1.MsgClawbackResponse) | Clawback removes the unvested tokens from a ClawbackVestingAccount. | GET|/canto/vesting/v1/tx/clawback|

 <!-- end services -->



<a name="canto/vesting/v1/vesting.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## canto/vesting/v1/vesting.proto



<a name="canto.vesting.v1.ClawbackVestingAccount"></a>

### ClawbackVestingAccount
ClawbackVestingAccount implements the VestingAccount interface. It provides
an account that can hold contributions subject to "lockup" (like a
PeriodicVestingAccount), or vesting which is subject to clawback
of unvested tokens, or a combination (tokens vest, but are still locked).


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `base_vesting_account` | [cosmos.vesting.v1beta1.BaseVestingAccount](#cosmos.vesting.v1beta1.BaseVestingAccount) |  | base_vesting_account implements the VestingAccount interface. It contains all the necessary fields needed for any vesting account implementation |
| `funder_address` | [string](#string) |  | funder_address specifies the account which can perform clawback |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | start_time defines the time at which the vesting period begins |
| `lockup_periods` | [cosmos.vesting.v1beta1.Period](#cosmos.vesting.v1beta1.Period) | repeated | lockup_periods defines the unlocking schedule relative to the start_time |
| `vesting_periods` | [cosmos.vesting.v1beta1.Period](#cosmos.vesting.v1beta1.Period) | repeated | vesting_periods defines the vesting schedule relative to the start_time |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |
