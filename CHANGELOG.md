## [0.63.8](https://github.com/pommee/goaway/compare/v0.63.7...v0.63.8) (2026-02-28)

### Bug Fixes

* **deps:** update dependencies ([83893cb](https://github.com/pommee/goaway/commit/83893cbe4d0bee323d29c88aebb67a17015eaba0))
* fixed quick action to allow/block a domain from the logs page ([0f8f6ee](https://github.com/pommee/goaway/commit/0f8f6eea4e342606dce4734cbe3f484b7cba92d1))

### UI/UX

* reworked design for client details modal ([1723dd4](https://github.com/pommee/goaway/commit/1723dd4d3942cddc70a07027dcb391d2d691ac1e))
* update logs page with some detail cards ([f466846](https://github.com/pommee/goaway/commit/f466846b7987573c4c13b827d7b0e66e3d013879))

## [0.63.7](https://github.com/pommee/goaway/compare/v0.63.6...v0.63.7) (2026-02-13)

### Bug Fixes

* **deps:** update client dependencies ([d013d92](https://github.com/pommee/goaway/commit/d013d92b794e7e38e52e20c1c0595ea163c33c46))
* **deps:** update to go version 1.26 ([9300add](https://github.com/pommee/goaway/commit/9300addf235ee1c28cee1f94262a7cdd98d94219))

## [0.63.6](https://github.com/pommee/goaway/compare/v0.63.5...v0.63.6) (2026-02-13)

### Bug Fixes

* use client filter for live query updates ([6b6d302](https://github.com/pommee/goaway/commit/6b6d30242f01d9a0086253b9a2cabed25ed3f614))

### Documentation

* update install documentation for proxmox ([5ce9d86](https://github.com/pommee/goaway/commit/5ce9d86d26689e54d957dc1c84616bf37d2ba675))

## [0.63.5](https://github.com/pommee/goaway/compare/v0.63.4...v0.63.5) (2026-01-18)

### Bug Fixes

* bump all client dependencies to prevent vulnerabilities ([1f27381](https://github.com/pommee/goaway/commit/1f273813145f345f9ade467ac9f8d908c54a307b))
* update golang version to 1.25.6 ([9636393](https://github.com/pommee/goaway/commit/9636393065de6c2bc49daef16833b742a19668e1))

### UI/UX

* correctly fetch all clients details when opening modal from logs page ([2adf424](https://github.com/pommee/goaway/commit/2adf42478a6e52042f723afd378da4a58ee038ed))
* update FrequencyChartBlockedDomains to use client name if possible ([213a5b4](https://github.com/pommee/goaway/commit/213a5b4bcd467ff3c37916d65580cc37bf4bb5e6))

## [0.63.4](https://github.com/pommee/goaway/compare/v0.63.3...v0.63.4) (2025-12-31)

### Bug Fixes

* **deps:** bump client dependencies ([f8d9a1d](https://github.com/pommee/goaway/commit/f8d9a1d2c89bc240b17778066a98072c1574e08d))

### UI/UX

* optimize ui for smaller viewports ([714f7ee](https://github.com/pommee/goaway/commit/714f7ee8359daeb77475a9b1dea2169a9d86191c))

## [0.63.3](https://github.com/pommee/goaway/compare/v0.63.2...v0.63.3) (2025-12-28)

### Bug Fixes

* added ability to change the name for any client ([2bca605](https://github.com/pommee/goaway/commit/2bca60597af60e2b9109653204d9a3eaeba81269))

## [0.63.2](https://github.com/pommee/goaway/compare/v0.63.1...v0.63.2) (2025-12-23)

### Bug Fixes

* let gorm handle createdat/updatedat fields ([8667a15](https://github.com/pommee/goaway/commit/8667a15db030dccdc3781e9a3feee1de89e16b2e))

## [0.63.1](https://github.com/pommee/goaway/compare/v0.63.0...v0.63.1) (2025-12-22)

### Bug Fixes

* fix bypass bug and cleanup client name/ip caches ([b1f2d68](https://github.com/pommee/goaway/commit/b1f2d6836268ca5755c060d74e4df477bcfa36f1))

## [0.63.0](https://github.com/pommee/goaway/compare/v0.62.25...v0.63.0) (2025-12-22)

### Features

* added ability to set 'bypass' for each client to bypass any rules ([1471ff1](https://github.com/pommee/goaway/commit/1471ff1d237ea49608dc399b0527fef6e1caf6d7))

### UI/UX

* added indicator to show blocking status along with a modal rework ([2d23669](https://github.com/pommee/goaway/commit/2d23669dd96606792413c1e50c29dfe986e66830))
* make it clear that 'client' is clickable in the logs view ([dc633be](https://github.com/pommee/goaway/commit/dc633becf4cd5d9be75087f0a6ead6eca2913990))

## [0.62.25](https://github.com/pommee/goaway/compare/v0.62.24...v0.62.25) (2025-12-20)

### Bug Fixes

* use net package instead of parsing raw strings for ip ([fb0b3bb](https://github.com/pommee/goaway/commit/fb0b3bb9a840bb69ad03a1781b542d97864b3996))

### Documentation

* add 'api.jwtSecret' to documentation ([b8eb1af](https://github.com/pommee/goaway/commit/b8eb1af46c1861a67100a2ac58ce7505af2ed753))

### UI/UX

* remove 'lucide-react' related and confetti library to minimize build. Also updated the rest of client dependencies ([918d4ba](https://github.com/pommee/goaway/commit/918d4ba8f02004b054b0e3623180a1bb97ec9d3d))

## [0.62.24](https://github.com/pommee/goaway/compare/v0.62.23...v0.62.24) (2025-12-11)

### Bug Fixes

* prevent initialization if previously done for blacklist ([c8d6ee5](https://github.com/pommee/goaway/commit/c8d6ee513ba56dd8c0c3593b0f248735bd7ebf98))

## [0.62.23](https://github.com/pommee/goaway/compare/v0.62.22...v0.62.23) (2025-12-08)

### Bug Fixes

* parse blocking time as number ([31706de](https://github.com/pommee/goaway/commit/31706de300c44c7af00aeadbafaa34cb10c9cab5))

## [0.62.22](https://github.com/pommee/goaway/compare/v0.62.21...v0.62.22) (2025-12-08)

### Performance Improvements

* better performance when fetching unique query types ([9042f1d](https://github.com/pommee/goaway/commit/9042f1d38847e828dd8a8b14ba78fadb8d8f06af))

### UI/UX

* added themes ([64aa4ad](https://github.com/pommee/goaway/commit/64aa4ad569e86dce42c49243aeef05b7279c356e))

## [0.62.21](https://github.com/pommee/goaway/compare/v0.62.20...v0.62.21) (2025-12-08)

### Bug Fixes

* add better toast for 'pause blocking' button in case of error ([16827a8](https://github.com/pommee/goaway/commit/16827a8b0762156da10b94861e0e1a43902aaaa2))
* remove 'white flash' for dark mode users ([aeb1894](https://github.com/pommee/goaway/commit/aeb18948c85981a44b47ede311ff1da0d17999a4))
* switched to only coloring log level instead of entire lines ([817dc24](https://github.com/pommee/goaway/commit/817dc245a4637cdfae0668f7111e1e02a1a56232))

### UI/UX

* add history tab to client card ([6565ec2](https://github.com/pommee/goaway/commit/6565ec2eea73e1d7fc451528c00d33db3305598b))
* added restart button in dashboard that will fully restart the entire application ([1d45174](https://github.com/pommee/goaway/commit/1d45174c8ce42c3b099f1e8129b0e4882deb0991))
* cursor-pointer on toggle switch ([32103ae](https://github.com/pommee/goaway/commit/32103ae2cb4f5c0628bae46230f7db73f858be6f))

## [0.62.20](https://github.com/pommee/goaway/compare/v0.62.19...v0.62.20) (2025-11-30)

### Bug Fixes

* remove port from default gateway ([ff476e4](https://github.com/pommee/goaway/commit/ff476e4d30987768b270d76002e1a9752ed7be84))
* save gateway value when saving from dashboard ([1958a57](https://github.com/pommee/goaway/commit/1958a5756fa9dcdaa2e8149a35345248ef704607))

## [0.62.19](https://github.com/pommee/goaway/compare/v0.62.18...v0.62.19) (2025-11-21)

### Bug Fixes

* dynamic jwt secret ([5769f87](https://github.com/pommee/goaway/commit/5769f8782b7453ca1c22a201b224b5ce48532f64))

### UI/UX

* allow wildcard on top level domain for resolution ([e4c7c08](https://github.com/pommee/goaway/commit/e4c7c08698ec988a8e42e4460dab3fbff34c8b0b))

## [0.62.18](https://github.com/pommee/goaway/compare/v0.62.17...v0.62.18) (2025-11-13)

### Bug Fixes

* prevent duplicated upstream queries ([caf6989](https://github.com/pommee/goaway/commit/caf698975800cd159d4cd222d7c254c492ac1a02))

### UI/UX

* allow IPv6 addr for upstream in ui ([f9b696f](https://github.com/pommee/goaway/commit/f9b696faa4828b520b50724f3927cc0c0b5e92f2))
* fix time since notification was created ([162a6af](https://github.com/pommee/goaway/commit/162a6af76c008528503c0b9ed4d3c9f3326c4c4b))

## [0.62.17](https://github.com/pommee/goaway/compare/v0.62.16...v0.62.17) (2025-11-10)

### Bug Fixes

* support wildcard in fqdn check ([788cdc9](https://github.com/pommee/goaway/commit/788cdc905bd0f73daff8e0f27ee82914221b7347))

## [0.62.16](https://github.com/pommee/goaway/compare/v0.62.15...v0.62.16) (2025-11-08)

### Bug Fixes

* return all upstreams including preferred ([5805526](https://github.com/pommee/goaway/commit/5805526205f296819b9671e565eb611b06b75d82))

### UI/UX

* outline variant for toggle section in settings ([2cdf96a](https://github.com/pommee/goaway/commit/2cdf96a66cca20ddc6375e8fd5a709fcf18a8b3a))
* wider client view by default ([8846fae](https://github.com/pommee/goaway/commit/8846fae40b8dea11abd0c970fafdc4e0490046e6))

## [0.62.15](https://github.com/pommee/goaway/compare/v0.62.14...v0.62.15) (2025-11-08)

### Bug Fixes

* use local gateway resolver for hostname ([daf7be2](https://github.com/pommee/goaway/commit/daf7be208b430442e9a1d368a819ae583bf9e6cc))

### UI/UX

* add gateway to settings page ([f5dc8d2](https://github.com/pommee/goaway/commit/f5dc8d2d39077a82bfd404c3f20440603cd7becd))

## [0.62.14](https://github.com/pommee/goaway/compare/v0.62.13...v0.62.14) (2025-11-08)

### Bug Fixes

* **deps:** bump golang version and client dependencies ([a966031](https://github.com/pommee/goaway/commit/a966031119a1a43320404c4493a03c3b00f7cfc9))
* restructure codebase, make setup and flow easier ([75fb86c](https://github.com/pommee/goaway/commit/75fb86cc00595a419c3d9035cc7bd6fd3c7b0936))

### Documentation

* add webpage for installation, configuration, setup and more ([00d0d21](https://github.com/pommee/goaway/commit/00d0d2115c98a41298ca27e93b158ae0923ca62c))

## [0.62.13](https://github.com/pommee/goaway/compare/v0.62.12...v0.62.13) (2025-11-08)

### Bug Fixes

* persist row size for log page and prevent server ip from using local unicast link ([c680dcc](https://github.com/pommee/goaway/commit/c680dcc4bfb69b8bfeab074b7ed6a566a94fc21d))

## [0.62.12](https://github.com/pommee/goaway/compare/v0.62.11...v0.62.12) (2025-10-31)

### Bug Fixes

* fix cache issue with whitelisted domains ([fe204f9](https://github.com/pommee/goaway/commit/fe204f914e682d4156cabb0e07f439e52075022c))

### Documentation

* add contributions page ([86d64bc](https://github.com/pommee/goaway/commit/86d64bc32b3a2d25689e6fecea22af86f44702ab))
* update contribution note ([57bbbe0](https://github.com/pommee/goaway/commit/57bbbe04d795c4706ed65c627aafc96bec51aefb))

### UI/UX

* add missing close handler to cancel button on PauseBlockingDialog ([835a1c2](https://github.com/pommee/goaway/commit/835a1c2fcd7eaef3c1abc53bc6b34612b24ea6b2))

## [0.62.11](https://github.com/pommee/goaway/compare/v0.62.10...v0.62.11) (2025-10-05)

### Bug Fixes

* switch go module for db import ([d6880b5](https://github.com/pommee/goaway/commit/d6880b50985c9e02747ae1ee62a4438bba471136))

## [0.62.10](https://github.com/pommee/goaway/compare/v0.62.9...v0.62.10) (2025-10-04)

### Bug Fixes

* correctly set port for upstream query using dot ([4dd21eb](https://github.com/pommee/goaway/commit/4dd21eb082a33d41a184aa71da4a218a38558ee7))

## [0.62.9](https://github.com/pommee/goaway/compare/v0.62.8...v0.62.9) (2025-10-02)

### Bug Fixes

* add retries when starting api server ([dc9aca9](https://github.com/pommee/goaway/commit/dc9aca97dc88c94bae66a3414adda77310eaea3e))
* respect port set for upstream over dot query ([ec08e44](https://github.com/pommee/goaway/commit/ec08e44fcbfeca91cdaab324638db86a77eb55bb))

## [0.62.8](https://github.com/pommee/goaway/compare/v0.62.7...v0.62.8) (2025-10-01)

### Bug Fixes

* added support for 'x-real-ip' header to doh queries ([f558337](https://github.com/pommee/goaway/commit/f558337f239e057c5e2a3be3e39f0c2becd0ae6b))

### UI/UX

* correct wording for certificates in settings ([a09638c](https://github.com/pommee/goaway/commit/a09638c0490209507871e20a53fa57c11a471de1))

## [0.62.7](https://github.com/pommee/goaway/compare/v0.62.6...v0.62.7) (2025-09-30)

### Performance Improvements

* various performance improvements to database queries ([6c23aa6](https://github.com/pommee/goaway/commit/6c23aa639b99f388672a73cf9fdf1ba07c48034a))

### UI/UX

* add 'page not found' and remove unused page transition ([95d0678](https://github.com/pommee/goaway/commit/95d0678a4f4d9fa2142e981b956ffd17ee8aceaa))

## [0.62.6](https://github.com/pommee/goaway/compare/v0.62.5...v0.62.6) (2025-09-29)

### Bug Fixes

* added ability to add multiple blacklists at once and various other fixes to the page ([c8dcdbf](https://github.com/pommee/goaway/commit/c8dcdbfdb8909d5898f8c8c4cc658beac730a30b))

### Documentation

* update readme with instructions on where to find logs and credentials for lxc ([b0022fe](https://github.com/pommee/goaway/commit/b0022fe48499a347653afaf2a4c03029db6dc0dd))

### UI/UX

* base64 encode list name when deleting to support special characters ([88764dc](https://github.com/pommee/goaway/commit/88764dc0d49b4c4774bb8469975569423ff89ece))

## [0.62.5](https://github.com/pommee/goaway/compare/v0.62.4...v0.62.5) (2025-09-29)

### Bug Fixes

* clean git state after semantic release ([fe5df12](https://github.com/pommee/goaway/commit/fe5df12c1b0c4cb3566a30d4a037de92a0d8cb7f))

## [0.62.4](https://github.com/pommee/goaway/compare/v0.62.3...v0.62.4) (2025-09-29)

### Bug Fixes

* correct behavior of unique constraint on blacklists table ([d9b064e](https://github.com/pommee/goaway/commit/d9b064e59f30ad3f267eda8e54cf00b58a02191b))

### Styles

* update to destructive button variant and fix pointer cursor case ([4025b8b](https://github.com/pommee/goaway/commit/4025b8bfe954d291ee2bd236c2b9fbcf69c155ad))
* update various buttons to align with global styling ([30ebcc9](https://github.com/pommee/goaway/commit/30ebcc9ed5e93638cfa659fa6bbfaf60a158ed1b))

### UI/UX

* add a test button for alert ([d6bc3e0](https://github.com/pommee/goaway/commit/d6bc3e0a2b8af493de4766290287cd93d54796d9))
* add padding at the bottom of audit log widget ([456c4eb](https://github.com/pommee/goaway/commit/456c4eb326429f631e7c6ac8b5df45c1dab4bacf))
* clean up resolution widgets ([f829585](https://github.com/pommee/goaway/commit/f829585518edd5f4fcd63e7f3d73dd9980b0fea2))
* fix refresh button colors for request timeline and response size timeline widgets ([ecb6db1](https://github.com/pommee/goaway/commit/ecb6db1c16e92f5eae60122cbb2a0b712307691b))
* generate quote for login page ([51b3a69](https://github.com/pommee/goaway/commit/51b3a694c36a17baffaf0ea6a2c9c8cff9ff8c04))
* show ip of client if name is unknown in clients map ([bff094f](https://github.com/pommee/goaway/commit/bff094f1640ff0a41337676885466a6b561c5af0))

## [0.62.3](https://github.com/pommee/goaway/compare/v0.62.2...v0.62.3) (2025-09-27)


### Bug Fixes

* set scheduled blacklist updates to true by default ([4cecb94](https://github.com/pommee/goaway/commit/4cecb942660828964b7d993ec7a89f18ac6daa80))
* use correct table when fetching domains for blacklist ([487c0ce](https://github.com/pommee/goaway/commit/487c0ce5ba365345509cb302c6fc9ed83370430e))

## [0.62.2](https://github.com/pommee/goaway/compare/v0.62.1...v0.62.2) (2025-09-25)


### Bug Fixes

* add gateway to settings ([4598b39](https://github.com/pommee/goaway/commit/4598b3998c8c3e0db7cc9f8afb168828a22b5e39))
* added local lookup of clients ([ab61a9a](https://github.com/pommee/goaway/commit/ab61a9aec7c616e3ad678ea78bbc7448670d4e6d))


### Performance Improvements

* faster loading of blacklists from database ([7a544cc](https://github.com/pommee/goaway/commit/7a544ccad60018a7d8016eaf67be3e1ced18bdd4))

## [0.62.1](https://github.com/pommee/goaway/compare/v0.62.0...v0.62.1) (2025-09-24)


### Bug Fixes

* add log updated logs ([0cc22b0](https://github.com/pommee/goaway/commit/0cc22b0f4e36070507d4e8e4523824ee7b5c0a74))
* add soa record type and default unhandled record type ([9209fa8](https://github.com/pommee/goaway/commit/9209fa840bfc4a6b4d1159c7a8efe17f294c6c26))
* add validation when adding a new prefetched domain ([f12adb9](https://github.com/pommee/goaway/commit/f12adb916ce087c81057d8fb736d2da245a793f3))
* added ability to test discord webhook for alerts ([324db21](https://github.com/pommee/goaway/commit/324db21a643b03dde78d35ae9afa7afc0a628207))
* database layer rewritten to use gorm ([2c9c832](https://github.com/pommee/goaway/commit/2c9c8325f2b8b1a71f33c709401d9ae46e052f4f))
* update server dependencies ([b4e2dd1](https://github.com/pommee/goaway/commit/b4e2dd1e244c388c768e01df69d8491af316ac82))
* update to golang 1.25.1 ([8a32e33](https://github.com/pommee/goaway/commit/8a32e33ffe54df7ed79b960788957ea9ee600dc3))

# [0.62.0](https://github.com/pommee/goaway/compare/v0.61.0...v0.62.0) (2025-09-05)


### Bug Fixes

* added futher logging ([eddf0ad](https://github.com/pommee/goaway/commit/eddf0ad24a4b06bfc09d07ac05f95b4f525219f7))
* smoother navigation bar ([9110c07](https://github.com/pommee/goaway/commit/9110c07a0b98764252fa036edba63cd01e3e8022))
* update client dependencies ([d4a349a](https://github.com/pommee/goaway/commit/d4a349ad0997263ba26444ba5e349b0bbdd5a428))
* upgrade go to 1.25 ([b38e210](https://github.com/pommee/goaway/commit/b38e21065bf75c3f419cb23f7b8b59dacb74e8dc))


### Features

* added alerts, currently only enabled for discord webhooks ([fa37c8d](https://github.com/pommee/goaway/commit/fa37c8d5c94cfe72de4c464c171674c6960da21c))

# [0.61.0](https://github.com/pommee/goaway/compare/v0.60.8...v0.61.0) (2025-08-22)


### Bug Fixes

* correct color vars for update custom list modal ([792c5f7](https://github.com/pommee/goaway/commit/792c5f731c6dbbe146e0ad3c41bf1050ff148be6))
* list url is now unique, requires old one to be removed and regenerated ([077de61](https://github.com/pommee/goaway/commit/077de613502bbad59118317c704d432d442ac277))
* ui improvements to the logs page when searching ([342d54a](https://github.com/pommee/goaway/commit/342d54ac61a474f5690ba0a85e6e932a740b43df))


### Features

* added ability to update list names, select multiple to then update or delete ([d8fc63d](https://github.com/pommee/goaway/commit/d8fc63d38c8c0459428482c031f21ca027b9fc11))

## [0.60.8](https://github.com/pommee/goaway/compare/v0.60.7...v0.60.8) (2025-08-12)


### Bug Fixes

* fixes to finding hostnames of clients ([b4a7a10](https://github.com/pommee/goaway/commit/b4a7a106e105467b5df70bed9c4440a80b0b0109))

## [0.60.7](https://github.com/pommee/goaway/compare/v0.60.6...v0.60.7) (2025-07-28)


### Bug Fixes

* add ping check for ws message before sending and set warnings to debug ([fb5ca94](https://github.com/pommee/goaway/commit/fb5ca94be08e2b7eb7811251e6e4fb16b8a05286))

## [0.60.6](https://github.com/pommee/goaway/compare/v0.60.5...v0.60.6) (2025-07-28)


### Bug Fixes

* improve 'export database' performance using chunk based streaming ([91bb777](https://github.com/pommee/goaway/commit/91bb777d1d89f469c620460c170eaf735a4c20c8))

## [0.60.5](https://github.com/pommee/goaway/compare/v0.60.4...v0.60.5) (2025-07-23)


### Bug Fixes

* improved the reverse hostname lookup process ([44cc6ff](https://github.com/pommee/goaway/commit/44cc6ff1301500ba6fb7347336e7a1e95085fa04))

## [0.60.4](https://github.com/pommee/goaway/compare/v0.60.3...v0.60.4) (2025-07-23)


### Bug Fixes

* reworked settings page and added more options ([552fc87](https://github.com/pommee/goaway/commit/552fc874ed9431de9c4fa90cadfb5c7c96a5e8d2))

## [0.60.3](https://github.com/pommee/goaway/compare/v0.60.2...v0.60.3) (2025-07-19)


### Bug Fixes

* simplify naming, has an impact on api responses ([25bd28e](https://github.com/pommee/goaway/commit/25bd28e4d3a59e6abf32f4dea4a1ae909b847213))

## [0.60.2](https://github.com/pommee/goaway/compare/v0.60.1...v0.60.2) (2025-07-19)


### Bug Fixes

* bump client dependencies ([b040443](https://github.com/pommee/goaway/commit/b040443789087e7b635d6b7691af18092607c052))
* bump go to 1.24.5 and bump backend dependencies ([bae75b1](https://github.com/pommee/goaway/commit/bae75b16aa5e12be188e2d573062b9ec4a332690))

## [0.60.1](https://github.com/pommee/goaway/compare/v0.60.0...v0.60.1) (2025-07-14)


### Bug Fixes

* fix issue related to port 853 always being used for upstream ([abead1f](https://github.com/pommee/goaway/commit/abead1f70cbc6982da9105b2fa3194f0ba4d3c1a))

# [0.60.0](https://github.com/pommee/goaway/compare/v0.59.0...v0.60.0) (2025-07-14)


### Features

* added reverse lookup for local clients ([60a022c](https://github.com/pommee/goaway/commit/60a022ca5cba7a54aa6b93edba88c141182eae45))

# [0.59.0](https://github.com/pommee/goaway/compare/v0.58.0...v0.59.0) (2025-07-14)


### Features

* added support for DoH (dns-over-https) ([e2d070a](https://github.com/pommee/goaway/commit/e2d070a7c741ad9330b722f9f868cafbceb69390))

# [0.58.0](https://github.com/pommee/goaway/compare/v0.57.0...v0.58.0) (2025-07-10)


### Features

* added support for DoT (DNS-over-TLS) ([0ef51c1](https://github.com/pommee/goaway/commit/0ef51c16214c4ce22e5536c45ea6eb1e78806533))

# [0.57.0](https://github.com/pommee/goaway/compare/v0.56.5...v0.57.0) (2025-07-06)


### Features

* added ability to toggle automatic blacklist updates at midnight ([9e09a6c](https://github.com/pommee/goaway/commit/9e09a6c852b06d994769e57d494fb36d5591768d))

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
