# CHANGELOG

## 0.1.2

- Add tests for querying feature flags

## 0.1.1

- `input` and `expected` values are always maps

This is to make parsing easier for statically typed languages (for now, Java)


## 0.2.4

* Added tests of provided by (env var) configs
  * missing environment variable
  * env var with string that can't be coerced to an integer

* Added tests of encrypted configs
  * including catching a decryption failure
* Test where the case of context property differs between rule and client context to show case-insensitivity
* Test that contexts without a key field are not shipped back to Prefab

## 0.2.4.1

* Small changes to the data formats in post.yaml for more consistency, easier implementation of the tests
