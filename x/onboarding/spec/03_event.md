<!--
order: 3
-->

# Event
The `x/onboarding` module emits the following event:

| Type       | Attribute Key      | Attribute Value               |
|:-----------|:-------------------|:------------------------------|
| onboarding | sender             | {senderBech32}                |
| onboarding | receiver           | {recipientBech32}             |
| onboarding | packet_src_channel | {packet.SourceChannel}        |
| onboarding | packet_src_port    | {packet.SourcePort}           |
| onboarding | packet_dst_port    | {packet.DestinationPort}      |
| onboarding | packet_dst_channel | {packet.DestinationChannel}   |
| onboarding | swap_amount        | {swappedAmount.String()}      |
| onboarding | convert_amount     | {convertCoin.Amount.String()} |
