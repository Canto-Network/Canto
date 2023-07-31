<!-- order: 5 -->

# BeginBlock

The begin block logic is executed at the end of each epoch.

## Cover Redelegation Penalty

- For all redelegation infos
  - get a redelegation entry of staking module 
  - calc diff between entry.SharesDst and dstDel.Shares
  - if calc is positive (meaning there is a penalty during the redelegation period)
    - calc penalty amt which is the token value of shares lost due to slashing 
    - send penalty amt of tokens to chunk 
    - chunk delegate additional shares corresponding penalty amt