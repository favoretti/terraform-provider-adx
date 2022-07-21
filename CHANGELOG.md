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
