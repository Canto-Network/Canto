<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# CHANGELOG

## Unreleased

### State Machine Breaking

- (deps) [#126](https://github.com/Canto-Network/Canto/pull/126) Bump Comsos-SDK to v0.50.6, CometBFT to v0.38.6, ibc-go to v8.2.1
  <!-- add ethermint bump up info after release -->

### Improvements

- (ante) [#126](https://github.com/Canto-Network/Canto/pull/126) Remove NewValidatorCommissionDecorator because its logic is duplicated with the logic implemented in the staking module's msg server.
- (x/*) [#126](https://github.com/Canto-Network/Canto/pull/126) Apply Cosmos-SDK improvements.
  - Remove `Type()` and `Route()` methods from all msgs
  - Remove `GetSigner()` methods from all msgs, move their logic to protobuf and define a custom GetSigner func if needed.
  - `authority` has been added to the required module to execute proposal msgs.

### Client Breaking

- (x/*) [#126](https://github.com/Canto-Network/Canto/pull/126) module-specific proposal and update params is moved to msg levelto to support msgs-based gov proposals.

<!-- Release links -->
