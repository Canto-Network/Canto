<!-- order: 5 -->

# BeginBlock

The begin block logic is executed at the end of each epoch.

## Cover Redelegation Penalty

- for all redelegation infos
  - get a redelegation entry of staking module 
  - calculate difference between `entry.SharesDst` and `dstDel.Shares`
  - if calculated value is positive (meaning there is a penalty during the redelegation period)
    - calculate penalty amount which is the token value of shares lost due to slashing 
    - send penalty amount of tokens to chunk 
    - chunk delegates additional shares corresponding penalty amount of tokens to validator