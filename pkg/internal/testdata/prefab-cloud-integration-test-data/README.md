# prefab-cloud-integration-test-data
Shared repository for integration test harness data


## Setup

### Environment Variables

* `PREFAB_INTEGRATION_TEST_API_KEY` : must be set to a valid api key for the Production environment of the integration test data organization
* `PREFAB_INTEGRATION_TEST_ENCRYPTION_KEY` : must be set to `c87ba22d8662282abe8a0e4651327b579cb64a454ab0f4c170b45b15f049a221`
* `NOT_A_NUMBER`: must be set to "not a number"
* `IS_A_NUMBER`: must be set to "1234"



# Updates

Versioning is managed via git tags and a version file now. Consumers of these examples use them as git submodules so they won't see updates automatically.

After committing changes, use the `tag_version.sh` script to automatically update both the version file and create a git tag with the same version