## [0.50.4](https://github.com/pommee/goaway/compare/v0.50.3...v0.50.4) (2025-06-07)


### Bug Fixes

* add wildcard for resolution ([c7e9558](https://github.com/pommee/goaway/commit/c7e9558eb4de2244f5c13137b88d3afa52dea2b2))

## [0.50.3](https://github.com/pommee/goaway/compare/v0.50.2...v0.50.3) (2025-06-06)


### Performance Improvements

* improve loading times of lists ([473dd23](https://github.com/pommee/goaway/commit/473dd2321861ffe90ee3c7b28b0edf31b89ea4a5))

## [0.50.2](https://github.com/pommee/goaway/compare/v0.50.1...v0.50.2) (2025-06-06)


### Bug Fixes

* query resolution before upstream ([79eb17e](https://github.com/pommee/goaway/commit/79eb17e822731f21f7b9887acf7d8b8c964c4a3d))

## [0.50.1](https://github.com/pommee/goaway/compare/v0.50.0...v0.50.1) (2025-06-06)


### Bug Fixes

* authentication turned on by default ([f56bb6e](https://github.com/pommee/goaway/commit/f56bb6e8db27e5e907594fe586ed510a3561b1ab))

# [0.50.0](https://github.com/pommee/goaway/compare/v0.49.10...v0.50.0) (2025-06-06)


### Features

* switch to alpine and add arm32 image ([26f8d00](https://github.com/pommee/goaway/commit/26f8d0045503d034f911d3df797e2cc341e05646))

## [0.49.10](https://github.com/pommee/goaway/compare/v0.49.9...v0.49.10) (2025-06-06)


### Bug Fixes

* bump client dependencies ([aac1800](https://github.com/pommee/goaway/commit/aac18008107a8807713a72397dd7feadbf8844ef))
* bump go version and dependencies ([030735c](https://github.com/pommee/goaway/commit/030735c09e8dc29cb06e78adc9785f47f0e7ec1c))

## [0.49.9](https://github.com/pommee/goaway/compare/v0.49.8...v0.49.9) (2025-06-06)


### Bug Fixes

* improve update process ([c15085d](https://github.com/pommee/goaway/commit/c15085d8f62745db647c61291d5f548cc6525073))

## [0.49.8](https://github.com/pommee/goaway/compare/v0.49.7...v0.49.8) (2025-06-06)


### Bug Fixes

* rework flags and remove remote pull of config as it is now created with defaults locally ([7c502da](https://github.com/pommee/goaway/commit/7c502daca18d2bafd7fe3026ebcf5048598a050c))

## [0.49.7](https://github.com/pommee/goaway/compare/v0.49.6...v0.49.7) (2025-06-02)


### Bug Fixes

* fixed arp parsing for windows ([1784a8a](https://github.com/pommee/goaway/commit/1784a8aaa0439e82a6065942625125845e447956))


### Performance Improvements

* faster parsing of domain name ([b9aeb18](https://github.com/pommee/goaway/commit/b9aeb18473a474801c8d6b6001bfcf076297a4d1))
* more efficient blacklist processing ([6b004ca](https://github.com/pommee/goaway/commit/6b004ca4b138d9733e18f31b48b95744b77f16a6))

## [0.49.6](https://github.com/pommee/goaway/compare/v0.49.5...v0.49.6) (2025-05-29)


### Bug Fixes

* shared pointer to config, fixes paused blocking ([a15acaa](https://github.com/pommee/goaway/commit/a15acaa9c456a72a71e349d6cd79f13cce4e9f1f))

## [0.49.5](https://github.com/pommee/goaway/compare/v0.49.4...v0.49.5) (2025-05-29)


### Bug Fixes

* correctly report used memory percentage ([bac6874](https://github.com/pommee/goaway/commit/bac687408bf5582dfe170a20694f34d9660e7003))

## [0.49.4](https://github.com/pommee/goaway/compare/v0.49.3...v0.49.4) (2025-05-29)


### Bug Fixes

* overall improvements to the query process ([3a29192](https://github.com/pommee/goaway/commit/3a2919293fe25b7c0ddf4f45004aca934d07c15c))

## [0.49.3](https://github.com/pommee/goaway/compare/v0.49.2...v0.49.3) (2025-05-28)


### Bug Fixes

* fully working network map in clients page ([b0ccc83](https://github.com/pommee/goaway/commit/b0ccc8331f788b50b1606bc33a9754d277762cca))

## [0.49.2](https://github.com/pommee/goaway/compare/v0.49.1...v0.49.2) (2025-05-26)


### Bug Fixes

* **deps:** bump client dependencies ([c347f8b](https://github.com/pommee/goaway/commit/c347f8b9c1b16612b84d05b90f07af4aba07c779))
* new db manager and tweaks to log saving process ([6f3cc6d](https://github.com/pommee/goaway/commit/6f3cc6d25ef468d4115cf2e69a52b40c6184ceab))

## [0.49.1](https://github.com/pommee/goaway/compare/v0.49.0...v0.49.1) (2025-05-25)


### Bug Fixes

* added interval selection to request timeline ([9ae7b26](https://github.com/pommee/goaway/commit/9ae7b267f5eb4299af6ce851c44a57ae9244a932))

# [0.49.0](https://github.com/pommee/goaway/compare/v0.48.10...v0.49.0) (2025-05-25)


### Features

* redesign of clients page to show live communication ([e58234c](https://github.com/pommee/goaway/commit/e58234c1ba335f233a7952faf2a1e16dc126f48b))

## [0.48.10](https://github.com/pommee/goaway/compare/v0.48.9...v0.48.10) (2025-05-25)


### Bug Fixes

* fix docker build command to fix pipeline ([e00c393](https://github.com/pommee/goaway/commit/e00c39348a67fda41b1261b4eebf408050d59350))

## [0.48.9](https://github.com/pommee/goaway/compare/v0.48.8...v0.48.9) (2025-05-25)


### Bug Fixes

* added ability to delete list ([832c5f7](https://github.com/pommee/goaway/commit/832c5f741a4666ac013c666e97777162947e4f43))
* populate blocklist cache once new list is added ([da4a614](https://github.com/pommee/goaway/commit/da4a614bf252d90d53305b4cc187fa2d3ebc979f))

## [0.48.8](https://github.com/pommee/goaway/compare/v0.48.7...v0.48.8) (2025-05-25)


### Bug Fixes

* better error handling for upstreams page and upstream pinger ([96f4b02](https://github.com/pommee/goaway/commit/96f4b02580803f7a02f5002bd2e3f913a4b0ca68))
* parse client last seen timestamp correctly ([ca39629](https://github.com/pommee/goaway/commit/ca39629ba04a0fede0ea89330525eb956afa8b71))

## [0.48.7](https://github.com/pommee/goaway/compare/v0.48.6...v0.48.7) (2025-05-24)


### Bug Fixes

* respect requested dashboard server ip ([1396c11](https://github.com/pommee/goaway/commit/1396c11bad798ec4ac1a8a4d869f932fda933b04))

## [0.48.6](https://github.com/pommee/goaway/compare/v0.48.5...v0.48.6) (2025-05-24)


### Bug Fixes

* respect set api port ([fc2dd02](https://github.com/pommee/goaway/commit/fc2dd02402fb940fe3a3dbfa60979a477f338e1c))

## [0.48.5](https://github.com/pommee/goaway/compare/v0.48.4...v0.48.5) (2025-05-24)


### Bug Fixes

* correctly pass on newest version ([59c9ab6](https://github.com/pommee/goaway/commit/59c9ab6006718741494323dee428dad53717365d))

## [0.48.4](https://github.com/pommee/goaway/compare/v0.48.3...v0.48.4) (2025-05-24)


### Bug Fixes

* versioned docker images ([8fe5c7d](https://github.com/pommee/goaway/commit/8fe5c7dd6b1fc2ab6d7e859490c76a113c07f588))

## [0.48.3](https://github.com/pommee/goaway/compare/v0.48.2...v0.48.3) (2025-05-24)


### Bug Fixes

* handle queries with no response from upstream ([57d7544](https://github.com/pommee/goaway/commit/57d75441a3f429557cebdd43202182f024907fcf))

## [0.48.2](https://github.com/pommee/goaway/compare/v0.48.1...v0.48.2) (2025-05-24)


### Bug Fixes

* parsing fix for timestamp ([27b5822](https://github.com/pommee/goaway/commit/27b58222a230028df159e60486229b405caba0b1))

## [0.48.1](https://github.com/pommee/goaway/compare/v0.48.0...v0.48.1) (2025-05-24)


### Bug Fixes

* correct order for release ([0c19ddd](https://github.com/pommee/goaway/commit/0c19ddd55dba66b2776872e04cd6e142f8a92901))

# [0.48.0](https://github.com/pommee/goaway/compare/v0.47.0...v0.48.0) (2025-05-24)


### Features

* new deployment strategy, versioned docker images and removed usage of cgo ([6d7bb00](https://github.com/pommee/goaway/commit/6d7bb0032b5a5c1aff1a62dfa8923b5e1c0ac6f2))
