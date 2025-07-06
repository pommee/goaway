## [0.56.5](https://github.com/pommee/goaway/compare/v0.56.4...v0.56.5) (2025-07-06)


### Bug Fixes

* added ability to remove and update multiple lists ([4dd2920](https://github.com/pommee/goaway/commit/4dd2920a68b546e438fcfc6374d72f29cf5dd2d6))

## [0.56.4](https://github.com/pommee/goaway/compare/v0.56.3...v0.56.4) (2025-06-30)


### Bug Fixes

* clearer indication when adding new list ([6fb1538](https://github.com/pommee/goaway/commit/6fb15381f3c337e624388fce7ce2506f3a5d1333))
* specify status for newly added list ([30048a5](https://github.com/pommee/goaway/commit/30048a5df04cb68a37d558548a3e0a8cb2b04a3f))
* update ui upon adding/removing an upstream ([f0a63dd](https://github.com/pommee/goaway/commit/f0a63dd9085016f0da660c6be883e0c3cb21b5f6))
* validate new upstream ip and port ([14d694f](https://github.com/pommee/goaway/commit/14d694f8ce67cd30fab3e73372cbb0259fa6e6c7))

## [0.56.3](https://github.com/pommee/goaway/compare/v0.56.2...v0.56.3) (2025-06-29)


### Bug Fixes

* reworked 'add list' modal and various other ui elements ([bdfa309](https://github.com/pommee/goaway/commit/bdfa309741115009a96b923403dc42999a580704))

## [0.56.2](https://github.com/pommee/goaway/compare/v0.56.1...v0.56.2) (2025-06-29)


### Bug Fixes

* env variables takes presence over settings file and flags ([328e6fc](https://github.com/pommee/goaway/commit/328e6fc3afeede68a200a4e310a1aefd3d3009e1))

## [0.56.1](https://github.com/pommee/goaway/compare/v0.56.0...v0.56.1) (2025-06-29)


### Performance Improvements

* performance improvement for resp size timeline and various ui changes ([6670ee1](https://github.com/pommee/goaway/commit/6670ee1b9ce9c984a7ec2d17c0cae15649db8042))

# [0.56.0](https://github.com/pommee/goaway/compare/v0.55.0...v0.56.0) (2025-06-28)


### Features

* added light/dark mode theme toggle ([1efad12](https://github.com/pommee/goaway/commit/1efad12f88ae6655c90fdb3dec3ca78934c58f5d))

# [0.55.0](https://github.com/pommee/goaway/compare/v0.54.7...v0.55.0) (2025-06-26)


### Features

* new response size timeline on the homepage ([873cb12](https://github.com/pommee/goaway/commit/873cb129189c41ad3b18ba97f7d2ed2404d1eff9))

## [0.54.7](https://github.com/pommee/goaway/compare/v0.54.6...v0.54.7) (2025-06-23)


### Bug Fixes

* add filters to clients page ([fdd9267](https://github.com/pommee/goaway/commit/fdd926718f5417eb0898c8c15770cd74b67b0205))

## [0.54.6](https://github.com/pommee/goaway/compare/v0.54.5...v0.54.6) (2025-06-20)


### Bug Fixes

* added sorting for certain log columns ([cb5ac49](https://github.com/pommee/goaway/commit/cb5ac492587abef3a7cf7d5a81a9ce757bb7e3e3))

## [0.54.5](https://github.com/pommee/goaway/compare/v0.54.4...v0.54.5) (2025-06-17)


### Bug Fixes

* trigger new release to get out previous changes ([9dcb829](https://github.com/pommee/goaway/commit/9dcb8292fb4065f7ef0da4f9f55aad0ecb46e9a8))

## [0.54.4](https://github.com/pommee/goaway/compare/v0.54.3...v0.54.4) (2025-06-17)


### Bug Fixes

* get initial list status when loading details ([9fddc3d](https://github.com/pommee/goaway/commit/9fddc3d491efcfbaadf3c18213cb65f894c83fd8))

## [0.54.3](https://github.com/pommee/goaway/compare/v0.54.2...v0.54.3) (2025-06-17)


### Bug Fixes

* better feedback when toggling, updating and removing a list ([8e61d0e](https://github.com/pommee/goaway/commit/8e61d0e3f35ccd63095551e7b0678d49ffc2a76c))
* better looking changelog ([515669f](https://github.com/pommee/goaway/commit/515669fa883e5bf68093a9d2d4ba6601b58caabc))
* hint that you will be logged out once password is changed ([3d80e70](https://github.com/pommee/goaway/commit/3d80e706beb085777796c0957a4c0e91ad5d5ea2))

## [0.54.2](https://github.com/pommee/goaway/compare/v0.54.1...v0.54.2) (2025-06-17)


### Bug Fixes

* always log ansi unless json is specified ([fe528dc](https://github.com/pommee/goaway/commit/fe528dcbe90fe0b9cf99fa8b9621fa61b677f3cc))
* respect false rate limit setting and warn when turned off ([ace8c4c](https://github.com/pommee/goaway/commit/ace8c4c22e88f367b72de4529841a9325872a417))

## [0.54.1](https://github.com/pommee/goaway/compare/v0.54.0...v0.54.1) (2025-06-17)


### Bug Fixes

* resolve 'overflows int' error for arm ([03d2a1c](https://github.com/pommee/goaway/commit/03d2a1c35d82b4cb132de72d04d00f04be61e0c6))

# [0.54.0](https://github.com/pommee/goaway/compare/v0.53.9...v0.54.0) (2025-06-17)


### Features

* added rate limit for login and generally more secure login flow ([d8ed524](https://github.com/pommee/goaway/commit/d8ed524136c21b8689d34c463d36768facf84d75))

## [0.53.9](https://github.com/pommee/goaway/compare/v0.53.8...v0.53.9) (2025-06-14)


### Bug Fixes

* added udpSize to config ([c7680fa](https://github.com/pommee/goaway/commit/c7680fa1c7db1a169a574180205e2b60db19b91f))
* always start dns server first ([e76afe4](https://github.com/pommee/goaway/commit/e76afe46ad6dd7b85e74c27021161dd31c097c18))

## [0.53.8](https://github.com/pommee/goaway/compare/v0.53.7...v0.53.8) (2025-06-14)


### Performance Improvements

* improve log loading performance by about 50x ([f6c4756](https://github.com/pommee/goaway/commit/f6c4756fb5faa8ca25f38f49879b2951ffb70182))

## [0.53.7](https://github.com/pommee/goaway/compare/v0.53.6...v0.53.7) (2025-06-13)


### Bug Fixes

* added import of database file ([7b83f85](https://github.com/pommee/goaway/commit/7b83f85af429591ef098c97163bf0a0d76282612))
* bump client dependencies ([63a6009](https://github.com/pommee/goaway/commit/63a6009058de82a3e0bca1477c7c2a3173658c14))

## [0.53.6](https://github.com/pommee/goaway/compare/v0.53.5...v0.53.6) (2025-06-13)


### Bug Fixes

* rw mutex for blacklist/whitelist ([95ac79a](https://github.com/pommee/goaway/commit/95ac79acadc6588f28888ee53eef1abbeb9683d0))

## [0.53.5](https://github.com/pommee/goaway/compare/v0.53.4...v0.53.5) (2025-06-13)


### Bug Fixes

* improve 'add new list' ui further ([2a70440](https://github.com/pommee/goaway/commit/2a704408fbeaead2f2c064d9417318c146ed6616))

## [0.53.4](https://github.com/pommee/goaway/compare/v0.53.3...v0.53.4) (2025-06-13)


### Bug Fixes

* improve lists page state handling and feedback ([17914e2](https://github.com/pommee/goaway/commit/17914e293e3d4809ebc1e2c5a9c5476762814657))

## [0.53.3](https://github.com/pommee/goaway/compare/v0.53.2...v0.53.3) (2025-06-13)


### Bug Fixes

* increase blacklist page load by ~40 times ([3c129fb](https://github.com/pommee/goaway/commit/3c129fb2299e5ea8ca002a1d8c5f6752da79a30a))

## [0.53.2](https://github.com/pommee/goaway/compare/v0.53.1...v0.53.2) (2025-06-12)


### Bug Fixes

* improve response ip and rtype, better ip view for logs, requires regeneration of database ([2fa0073](https://github.com/pommee/goaway/commit/2fa0073a62ae3b0e57adada2e6208fead370ed8a))

## [0.53.1](https://github.com/pommee/goaway/compare/v0.53.0...v0.53.1) (2025-06-11)


### Bug Fixes

* remove appuser ([9690755](https://github.com/pommee/goaway/commit/9690755322fcd77dd71f2ddaf2073afff2d5ce51))

# [0.53.0](https://github.com/pommee/goaway/compare/v0.52.1...v0.53.0) (2025-06-11)


### Bug Fixes

* improve volume mounts and dev setup ([fc8536f](https://github.com/pommee/goaway/commit/fc8536ff74cd2f40a2b0413d31ce8440008ee032))


### Features

* make in-app updates optional, false by default ([d51218c](https://github.com/pommee/goaway/commit/d51218c91d728b1b8a41c9f2a9ccef749df4209d))

## [0.52.1](https://github.com/pommee/goaway/compare/v0.52.0...v0.52.1) (2025-06-11)


### Bug Fixes

* take SQLite WAL mode into consideration when calculating DB size and exporting backup file ([16b56cf](https://github.com/pommee/goaway/commit/16b56cfa20e7868a2dca08493ad574423904a0a9))

# [0.52.0](https://github.com/pommee/goaway/compare/v0.51.1...v0.52.0) (2025-06-10)


### Bug Fixes

* increase arp lookup time since this can take longer on a bigger network ([9cf0e97](https://github.com/pommee/goaway/commit/9cf0e976d2a1bb4c5b8111e838e23e3ba6018f7e))
* log and return an error when loading a blacklist or whitelist fails. ([1943c51](https://github.com/pommee/goaway/commit/1943c51858ca3f4abd27f5b0490acca06e8f26c1))
* make domain unique instead of IP and clear cache (issue 23) ([002569c](https://github.com/pommee/goaway/commit/002569c489cf764d0e6fcb7f6323ffb75dca0d26))
* return 0 if temperature can't be read to reduce error logs ([6a71a30](https://github.com/pommee/goaway/commit/6a71a3035101c51fdb07e9f5b9a40e5ca7678d36))


### Features

* allow binding to a specific IP ([10bab26](https://github.com/pommee/goaway/commit/10bab2642be84d82ca817c8cdac0fb23a33a8c77))

## [0.51.1](https://github.com/pommee/goaway/compare/v0.51.0...v0.51.1) (2025-06-08)


### Bug Fixes

* token improvements, dont refresh upon each request ([72dc0f2](https://github.com/pommee/goaway/commit/72dc0f20c1cba39c3896d05ee484893bffc49cc6))

# [0.51.0](https://github.com/pommee/goaway/compare/v0.50.5...v0.51.0) (2025-06-08)


### Features

* added whitelist page ([d91ea7c](https://github.com/pommee/goaway/commit/d91ea7c0d4870cdb22562a20c0cd111204115035))

## [0.50.5](https://github.com/pommee/goaway/compare/v0.50.4...v0.50.5) (2025-06-07)


### Bug Fixes

* support deeply nested subdomains in wildcard resolution ([1ca3140](https://github.com/pommee/goaway/commit/1ca3140fadcab16733808b8ea5261446966f1638))

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
