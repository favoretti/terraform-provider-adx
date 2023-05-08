## v0.0.22

* `adx_function` supports imports now (#48)

## v0.0.21

* Add concurrency and maxSourceRecordsForSingleIngest as optional params to MVs (#42)
* add concurrency and maxSourceRecordsForSingleIngest as optional parameter to mv
* add new predicate and augmented tests

## v0.0.20

* Fixed issue #35. Async materialized view creation now properly polls ADX operations to wait for completion
* Documentation fix for `adx_table_update_policy` attribute `source_table`
* Fixed adding columns to a table using `column` method (previously it only worked using `table_schema` method)
* Fixed Issue #33 "adx_table: Column always wants to be updated" by using DiffSuppression

## v0.0.19

* Fixed issue #30. Float parsing issues for default and empty values from ADX for streamingingestion policy
* Fixed #32. Add `apacheavro` to table_mapping valid types

## v0.0.18

* Added folder and docstring parameters to functions, materialized views, and tables
* Adding streamingingestion policy resource for tables

## v0.0.17

* Bug fixes for tables and views "with" attributes when resource is updated (some are incompatible with .alter)
* Bug fix for cache policies on follower databases (API is eventually consistent, had to add polling for state read)
* Added TF import functionality to all resources
* Acceptance tests for resource updates

## v0.0.16

* Due to a glitch in time-space continuum, this is identical to v0.0.15

## v0.0.15

### Caching polices for tables and views in follower databases

* Added acceptance tests for caching policies on materialized views
* `follower_database` flag for both table and mv caching policies

### Materialized View Improvements

* Acceptance tests for materialized views
* Materialized view bug fixes (found by new acceptance tests)
* allow_mv_without_rls flag for both rls & mv policy creation to help with common issues creating views on top of tables configured with RLS

## v0.0.14

* added csv, orc, parquet mapping types
* fixed cache concurrecy issue, fixed bug in hasStatementResults

## v0.0.13

* Added retention, caching, and RLS policies for materialized views

`adx_materialized_view_caching_policy`
`adx_materialized_view_retention_policy`
`adx_materialized_view_row_level_security_policy`

* Input validation bug fixes for mapping and function resources
* Fixed cluster uri param name in docs
* Fixed possible duplication of function resources

## v0.0.12

* Resource IDs of `adx_table` and `adx_table_mapping` were changed to match the structure of the newly added resources. Should have no impact on normal operation.
* Lazy Init of provider to resolve Issue Unable to create adx and then use the provider at the same time #2
* Cluster config per resource (to manage resources across multiple clusters)
* Refactoring of client & query helpers
* Checks for deleted resources in ADX (previously this caused errors and this provider would not recognize to re-create them)
* Client caching for optimizing control of many resources across many clusters (client per cluster)

## v0.0.11

* New resource: `adx_table_caching_policy`
* New resource: `adx_table_partitioning_policy`
* New resource: `adx_materialized_view`

## v0.0.10

### Table resource improvements:

* Ability to update table definitions (.alter & .alter-merge),
* Table creation from query (.set, .set-or-replace etc..)

### New resources:

* User defined functions
* Table row level security policy
* Table batch ingestion policy
* Table Retention policy
* Table update policy
* Added helper methods for improving ID generation & maintaining policy objects (to make adding more policy types easier)

Table and table mapping resources were not updated to use the new id generation since it requires state migration

Upgraded to go 1.18 and terraform sdk v2.8.0

## v0.0.9

* Fix crash mentioned in #3

## v0.0.8

* Fix validation bugs for `adx_table`

## v0.0.7

* Do not recreate resources as updating them seems to be supported and work well

## v0.0.6

* Make `table_schema` and `column` definition formats in `adx_table` interchangeable

## v0.0.5

* Add support for HCL-style table definitions

## v0.0.4

* Fix some typos (trigger-happy on releases FTW)

## v0.0.3

* Add initial documentation

## v0.0.2

* Add GitHub Actions and TF Registry release

## v0.0.1

* Initial working version
