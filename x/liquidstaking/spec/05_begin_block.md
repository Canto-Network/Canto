<!-- order: 5 -->

# BeginBlock

The end block logic is executed at the end of each epoch.

## Cover Redelegation Penalty

- For all redelegation infos
  - get a redelegation entry of staking module 
  - calc diff between entry.SharesDst and dstDel.Shares
  - if calc is positive (meaning there is a penalty during the redelegation period)
    - calc penalty amt: `dstVal.TokensFromShares(diff).Ceil().TruncateInt` which is the token value of shares lost due to slashing 
    - if penalty amt is bigger than unpairing insurance balance
      - mark penalty amt to re-delegation info, so that it can be handled in handlePairedInsurance.
      - return
    - send penalty amt of tokens to chunk (if unpairing insurance balance is below penalty amt, send all insurance's balance to chunk)
    - chunk delegate additional shares corresponding penalty amt