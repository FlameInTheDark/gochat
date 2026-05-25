# [1.15.0](https://github.com/FlameInTheDark/gochat/compare/v1.14.0...v1.15.0) (2026-05-25)


### Bug Fixes

* dm call functionality fix ([568699c](https://github.com/FlameInTheDark/gochat/commit/568699c83a759a6459c51519c5e990693ac9cc2f))
* last channel message id on deleted messages fix ([9e9dc4c](https://github.com/FlameInTheDark/gochat/commit/9e9dc4c1a0f097327deef8c8261d792522012ceb))
* lint fixes ([164d6d7](https://github.com/FlameInTheDark/gochat/commit/164d6d755f0b33b605dc54bb8d5b6e163b09198c))
* readstate error on ack ([1046616](https://github.com/FlameInTheDark/gochat/commit/104661619e663c9b527c5d4bd30e70bf6ac325a2))
* s3 upload fix for multipart upload ([0fa0f66](https://github.com/FlameInTheDark/gochat/commit/0fa0f66886b840dc27898a2cd508953ce2e8147d))
* unread state issue ([2be9ffd](https://github.com/FlameInTheDark/gochat/commit/2be9ffdae52f9d39c636148c4c7f21dab8efe9d4))
* WS stability improvement ([be46d7f](https://github.com/FlameInTheDark/gochat/commit/be46d7ff9913d2f2f7b93630afba948233f5ad85))


### Features

* database connector with yugabyte/pgx driver ([40347e2](https://github.com/FlameInTheDark/gochat/commit/40347e28389749c879e7691e2451db2c7ceb58a2))
* profile banners and personal notes ([b88745a](https://github.com/FlameInTheDark/gochat/commit/b88745a77da0dfb7025466b8e9f96a53bba30347))
* YugabyteDB ([88dbe19](https://github.com/FlameInTheDark/gochat/commit/88dbe19dba2976904cc4877e1b160f647302fd0f))
* YugabyteDB migration ([0d85a10](https://github.com/FlameInTheDark/gochat/commit/0d85a1092a84d42d6789523458cd5d219f4c942e))

# [1.14.0](https://github.com/FlameInTheDark/gochat/compare/v1.13.0...v1.14.0) (2026-05-14)


### Bug Fixes

* voice region selection fix ([e335fb8](https://github.com/FlameInTheDark/gochat/commit/e335fb829ef90f3f437937d24b686a6338f153e7))


### Features

* DM calls feature ([7506b0d](https://github.com/FlameInTheDark/gochat/commit/7506b0d8169e4ef36082aa313f6253bd1f57bb11))
* github actions for search service ([2a94255](https://github.com/FlameInTheDark/gochat/commit/2a9425599e5a9578323df80b0e3f8afa200955b9))
* guild discovery ([27fa1ae](https://github.com/FlameInTheDark/gochat/commit/27fa1aea0412c2897591cb33828ee16a0835fd4c))
* roles and threads improvement ([d74fae5](https://github.com/FlameInTheDark/gochat/commit/d74fae5e9184faebe78058457053d26032291a09))
* thread creation fix for deleted threads and non-message threads ([978f50d](https://github.com/FlameInTheDark/gochat/commit/978f50d363ef711d828e9ad0806a81355d76e537))
* UX improvement and fixes ([86097ae](https://github.com/FlameInTheDark/gochat/commit/86097aed7d616013dda908e5b45fca008d2949a9))

# [1.13.0](https://github.com/FlameInTheDark/gochat/compare/v1.12.0...v1.13.0) (2026-04-30)


### Bug Fixes

* security audit and fixes ([ee586db](https://github.com/FlameInTheDark/gochat/commit/ee586db6ac14e78756e3a6ff542b6528e50524dd))


### Features

* DTLS for SFU WebRTC ([ab955cf](https://github.com/FlameInTheDark/gochat/commit/ab955cfc96a66efc44e8f26838d19219000ef6f4))
* streaming service ([fb6bd8c](https://github.com/FlameInTheDark/gochat/commit/fb6bd8c789db435eb8123c7d438307655cc2ec20))
* switched from local DAVE implementation to external library usage ([b6a406e](https://github.com/FlameInTheDark/gochat/commit/b6a406efad506f6130d8a055a39bfd7f242148ed))
* system notifications channel fixes ([d4bff6f](https://github.com/FlameInTheDark/gochat/commit/d4bff6f17126db92105707afe3c4c7aad326c72b))
* Updated stream connection flow and video stream stability improvement ([1028db8](https://github.com/FlameInTheDark/gochat/commit/1028db8dbae41a6b6fa5513c560f4cfc0d964ba1))

# [1.12.0](https://github.com/FlameInTheDark/gochat/compare/v1.11.0...v1.12.0) (2026-03-31)


### Bug Fixes

* 500 error on expired invites fix ([957c7b6](https://github.com/FlameInTheDark/gochat/commit/957c7b6571271e130366eb3de3e4c56fab69f60f))
* fixed broken test ([7702485](https://github.com/FlameInTheDark/gochat/commit/77024857bcc7707ac890e59cb2ce9855054e83c9))
* lint fixes ([a68d8e4](https://github.com/FlameInTheDark/gochat/commit/a68d8e4580c042a3fa5f5b73ae4bb4dfd550bf7e))
* password recovery fix ([de84874](https://github.com/FlameInTheDark/gochat/commit/de8487444b3fe3a29d399ae5e3144ddd21f110ea))
* password recovery id fix ([bfede73](https://github.com/FlameInTheDark/gochat/commit/bfede734fc6efe99021fe24f5a9c44cda1fbc0cb))
* race condition and CI fixes ([d480ce6](https://github.com/FlameInTheDark/gochat/commit/d480ce6f9b61f646e782aa6ebd19aeb714c5d64f))


### Features

* proactive cache ([7a65f3f](https://github.com/FlameInTheDark/gochat/commit/7a65f3f03c9fc16dc9c7975988e110bc5eda0903))
* refactoring and minor issues fix ([8541923](https://github.com/FlameInTheDark/gochat/commit/85419230c96c6b2e972317b5c171456fe75df48f))

# [1.11.0](https://github.com/FlameInTheDark/gochat/compare/v1.10.0...v1.11.0) (2026-03-30)


### Bug Fixes

* log levels for traces, prevent event spam with info messages ([a80097e](https://github.com/FlameInTheDark/gochat/commit/a80097e46f58e6a7a1729f6e31ceb44ea6172e31))
* log levels, prevent event spam with info messages ([3705a5a](https://github.com/FlameInTheDark/gochat/commit/3705a5a5846d8733f2ee9453653e458aeb2b7652))
* logging all as error fix ([cef47c6](https://github.com/FlameInTheDark/gochat/commit/cef47c66abad63d45b56c8fbc24bf6e23edebb11))
* optimization and fixes ([0d0d1c6](https://github.com/FlameInTheDark/gochat/commit/0d0d1c6a0656fcec8f9de3064dcaeaf2d409ec75))
* reduced metrics spam ([33fbf04](https://github.com/FlameInTheDark/gochat/commit/33fbf04f35ffc8f16cc6a5e9840e79398c9a2ec3))
* SFU config update for otlp ([797cd06](https://github.com/FlameInTheDark/gochat/commit/797cd06b74655237a40f61dd86879e51a37cdaab))
* SFU flow optimization ([3c5b0a4](https://github.com/FlameInTheDark/gochat/commit/3c5b0a4c868ffd0b752a04121cdf0a27db4e60ed))
* SFU public ip for NAT 1to1 support ([2630f98](https://github.com/FlameInTheDark/gochat/commit/2630f98d4b7b71783b117d43682ba6da74d9443c))
* SFU video quality fix ([67f0e18](https://github.com/FlameInTheDark/gochat/commit/67f0e18a7b22b1dd05059933e6126d6dbbc6fb10))


### Features

* client rebuild ([ae24b44](https://github.com/FlameInTheDark/gochat/commit/ae24b4404c7059def0bca7ddf7ad26ec4a571823))
* device aware settings ([fe8f3bf](https://github.com/FlameInTheDark/gochat/commit/fe8f3bfb1fcd4fb57762df7e1abd1e51f65a72df))
* E2EE and fixes ([dd2c72e](https://github.com/FlameInTheDark/gochat/commit/dd2c72e03db1d0167ce1ab5a2743b008df68e313))
* limit for device settings list ([47d6d38](https://github.com/FlameInTheDark/gochat/commit/47d6d38120c0606123a60d5b7cd3156e3be4e57e))
* new clients with totp ([1402455](https://github.com/FlameInTheDark/gochat/commit/14024557b9dbda99172c5a5646c8e88169656ae6))
* SFU port range configuration ([44e0c55](https://github.com/FlameInTheDark/gochat/commit/44e0c55adb14579df39084ac300a7b19a46cc5bf))
* SFU v2 and DAVE integration documentation ([055f991](https://github.com/FlameInTheDark/gochat/commit/055f9917d24a0dc23c02550de6daf94333a4b13f))
* SFU v2 protocol with DAVE ([41cd390](https://github.com/FlameInTheDark/gochat/commit/41cd390ca522c2a561a3520e7784f8ee29585bf1))
* TOPT authentication security ([32a296c](https://github.com/FlameInTheDark/gochat/commit/32a296ce02d8b95df270dc28f212b02f5af3deaf))

# [1.10.0](https://github.com/FlameInTheDark/gochat/compare/v1.9.0...v1.10.0) (2026-03-25)


### Bug Fixes

* added fallback for failed webP preview generation ([f919bba](https://github.com/FlameInTheDark/gochat/commit/f919bba9e3863a7e7cf0fe2127a969bd518faaae))
* added fallback frame extraction and rendering for failed animated webP preview generation ([19c7b71](https://github.com/FlameInTheDark/gochat/commit/19c7b7174d5cc0dbc9447a0bac686758286f174f))
* cache and messages optimization ([e873ba8](https://github.com/FlameInTheDark/gochat/commit/e873ba83b4d1472df804ba9b4854dfbb43f45bc5))
* cache batching optimization ([0a1f9a1](https://github.com/FlameInTheDark/gochat/commit/0a1f9a1a6583140d8ea2245a321e3482007be25c))
* channels cache fix ([c0a726d](https://github.com/FlameInTheDark/gochat/commit/c0a726da5c98d752e20e4bed508ecbf14ec4118e))
* fixed auth confirmation parsing and errors observability ([5047986](https://github.com/FlameInTheDark/gochat/commit/5047986349bf75bdd8528fd5ac533412989b5e02))
* fixed channel creation ([e5aa6bd](https://github.com/FlameInTheDark/gochat/commit/e5aa6bd84e2269e0d387033226ca5c56ea864607))
* fixed channel id parsing ([25b4e29](https://github.com/FlameInTheDark/gochat/commit/25b4e2924e74eb35c104819fab24322cdcf1e4cc))
* fixed channel reorder ([85f578f](https://github.com/FlameInTheDark/gochat/commit/85f578f48c628af170aa1ee3f8d7b67d1500eeda))
* fixed friend request fail on second sending ([69d30c8](https://github.com/FlameInTheDark/gochat/commit/69d30c8a727d8262bc3906507a632ed8bba0eabd))
* fixed user settings user id parsing ([04f6202](https://github.com/FlameInTheDark/gochat/commit/04f6202b8c8a1be961bc5894986f3cc5eba93476))
* more cache batching optimization ([dadd369](https://github.com/FlameInTheDark/gochat/commit/dadd369cca1abbad2e46a6556b20363c54608052))
* websocket close error handling ([951cc88](https://github.com/FlameInTheDark/gochat/commit/951cc88ea7d5dfbe18ac2a95bc5a8720bdc2591e))


### Features

* added more notification options in user settings ([b95fc7c](https://github.com/FlameInTheDark/gochat/commit/b95fc7cf8c90a8146d2fa61087824f0f528e05f2))
* emoji info route for UI tooltips ([6f882fa](https://github.com/FlameInTheDark/gochat/commit/6f882fa72df269b31bca6233328c016b3fc0ca11))
* message reactions ([79312ab](https://github.com/FlameInTheDark/gochat/commit/79312abed558c477781afb673bb29ba30d2eb24d))
* migration image for easy database migrations ([6ce96e3](https://github.com/FlameInTheDark/gochat/commit/6ce96e39463b351abf6825c39cf2c84605e76d73))

# [1.9.0](https://github.com/FlameInTheDark/gochat/compare/v1.8.0...v1.9.0) (2026-03-20)


### Features

* added observability features (OpenObserve and OpenTelemetry), fixed SFU video renegotiation, fixed migrations ([aa82431](https://github.com/FlameInTheDark/gochat/commit/aa824310a17119eafcbd5fa4d9526abdf229763f))

# [1.8.0](https://github.com/FlameInTheDark/gochat/compare/v1.7.0...v1.8.0) (2026-03-19)


### Features

* additional member customization ([e5a20f2](https://github.com/FlameInTheDark/gochat/commit/e5a20f2ad70ac9ac674cd0bfcd3650e878fb1896))
* threads and replies ([980eae3](https://github.com/FlameInTheDark/gochat/commit/980eae38318efe4f7c816e06f87c50bc678a634d))

# [1.7.0](https://github.com/FlameInTheDark/gochat/compare/v1.6.0...v1.7.0) (2026-03-09)


### Bug Fixes

* SFU service crash fix on user join voice channel. Increased limit for user audio input and output parameters in settings ([6aa89c6](https://github.com/FlameInTheDark/gochat/commit/6aa89c68212840005aee344a3063126c5a94160d))


### Features

* added embedder to github actions ([815f683](https://github.com/FlameInTheDark/gochat/commit/815f68318ef2438b95c009186deae806e0f0d7a7))
* added moderation features kick/ban/unban ([a01a522](https://github.com/FlameInTheDark/gochat/commit/a01a5226f674bf28cb5ab1fc51778118bf1c46a8))
* attachment service refactoring ([46b933d](https://github.com/FlameInTheDark/gochat/commit/46b933d8e9f43206f7efbb5613dee727f83ec521))
* custom server emojis ([57abacb](https://github.com/FlameInTheDark/gochat/commit/57abacb1b667678f38bf08072d252762bd414bcf))
* DM channel message search ([83a3f3a](https://github.com/FlameInTheDark/gochat/commit/83a3f3a9b4d268fd04ad631a393b5a8a673cdb91))
* improved embedder service parsing for twitter/x ([924db96](https://github.com/FlameInTheDark/gochat/commit/924db9668f6f7f5c7a2c6237f2c14d2858f211dd))
* improved embedder with cache, better thumbnails and exclusion regex patterns configuration ([a357854](https://github.com/FlameInTheDark/gochat/commit/a35785441584eb992780c5d161daf0095de4f520))
* message embed and url embed generator ([0d2d186](https://github.com/FlameInTheDark/gochat/commit/0d2d1866a434e85427fef6bce82aa5c1a442d745))
* minor improvements and documentation ([776de89](https://github.com/FlameInTheDark/gochat/commit/776de89d45ef23f0b1c06112de402ccedb000d9f))
* optimizations and bugfixes ([cda620c](https://github.com/FlameInTheDark/gochat/commit/cda620c2cd24d0f24c5c9c4406c4f4423b5f389e))
* optimizations and bugfixes, documentation ([03a7cdf](https://github.com/FlameInTheDark/gochat/commit/03a7cdfcd85c376fac9ce3fc606ff45d89ddf05c))
* role ordering ([0d64393](https://github.com/FlameInTheDark/gochat/commit/0d643932f202459ee21203d36c4efae926b5e9d9))
* server custom emojis ([76298da](https://github.com/FlameInTheDark/gochat/commit/76298daae4468cd0ddeba0c0e9c7aac6f1a3dc26))

# [1.6.0](https://github.com/FlameInTheDark/gochat/compare/v1.5.0...v1.6.0) (2025-10-29)


### Features

* added DashaMail transactional email provider ([bf10891](https://github.com/FlameInTheDark/gochat/commit/bf108917c74047f55748e1aca24892cf505b4587))

# [1.5.0](https://github.com/FlameInTheDark/gochat/compare/v1.4.0...v1.5.0) (2025-10-29)


### Features

* added attachments service that handle attachments upload and preview creation ([db1e439](https://github.com/FlameInTheDark/gochat/commit/db1e4398f3a4b1958eb64adbd706af6fb9e65480))
* added attachments uploading service and managing routes for icons/avatars, added guild deletion route ([b1ea3d5](https://github.com/FlameInTheDark/gochat/commit/b1ea3d515958ed8284c5f92185de972ac8752202))
* added DMs and file upload ([d823320](https://github.com/FlameInTheDark/gochat/commit/d823320e0acaec235f9181175b8214fb2a1f309f))
* added read states, changed user settings ([70323e1](https://github.com/FlameInTheDark/gochat/commit/70323e143373905a38e4a61f84676f1d155a0e63))
* added user presence statuses ([8de710e](https://github.com/FlameInTheDark/gochat/commit/8de710ec28e8a92a0bd177d4f27eccd114b3a913))
* mention notifications and video calls ([a3bcbdf](https://github.com/FlameInTheDark/gochat/commit/a3bcbdf27d9539cee9000c72473ef07febbc1653))
* SFU voice server ([5342ac6](https://github.com/FlameInTheDark/gochat/commit/5342ac6f2cb700a56d07b7af75c65804051d8961))
* system messages and user typing ([046b041](https://github.com/FlameInTheDark/gochat/commit/046b0416b95bdce945c9c5b0fb628a62aaba8582))

# [1.5.0](https://github.com/FlameInTheDark/gochat/compare/v1.4.0...v1.5.0) (2025-10-29)


### Features

* added attachments service that handle attachments upload and preview creation ([db1e439](https://github.com/FlameInTheDark/gochat/commit/db1e4398f3a4b1958eb64adbd706af6fb9e65480))
* added attachments uploading service and managing routes for icons/avatars, added guild deletion route ([b1ea3d5](https://github.com/FlameInTheDark/gochat/commit/b1ea3d515958ed8284c5f92185de972ac8752202))
* added DMs and file upload ([d823320](https://github.com/FlameInTheDark/gochat/commit/d823320e0acaec235f9181175b8214fb2a1f309f))
* added read states, changed user settings ([70323e1](https://github.com/FlameInTheDark/gochat/commit/70323e143373905a38e4a61f84676f1d155a0e63))
* added user presence statuses ([8de710e](https://github.com/FlameInTheDark/gochat/commit/8de710ec28e8a92a0bd177d4f27eccd114b3a913))
* mention notifications and video calls ([a3bcbdf](https://github.com/FlameInTheDark/gochat/commit/a3bcbdf27d9539cee9000c72473ef07febbc1653))
* SFU voice server ([5342ac6](https://github.com/FlameInTheDark/gochat/commit/5342ac6f2cb700a56d07b7af75c65804051d8961))
* system messages and user typing ([046b041](https://github.com/FlameInTheDark/gochat/commit/046b0416b95bdce945c9c5b0fb628a62aaba8582))

# [1.4.0](https://github.com/FlameInTheDark/gochat/compare/v1.3.0...v1.4.0) (2025-10-04)


### Features

* added ability to fetch messages around ([133ad34](https://github.com/FlameInTheDark/gochat/commit/133ad3419578b8cb0e86bb8b0c84f37976b0330d))
* added user settings with read state, some metrics ([f82a3c5](https://github.com/FlameInTheDark/gochat/commit/f82a3c5a162e380df0386067d923755b3fbfe78d))
* transactions for guild channel creation and removal operations, channel ordering on creation ([8663ed5](https://github.com/FlameInTheDark/gochat/commit/8663ed50992f5a4d6a931433f578ac828340bb72))
* websocket events and guild api improvements ([fe434c2](https://github.com/FlameInTheDark/gochat/commit/fe434c2cdfa0937b6b3f54a2534227d8f0590134))

# [1.3.0](https://github.com/FlameInTheDark/gochat/compare/v1.2.0...v1.3.0) (2025-09-22)


### Features

* a little speedup of rate limiting ([8155d4e](https://github.com/FlameInTheDark/gochat/commit/8155d4e43340cfe6e83a24c2d9aad477ef0c882f))
* role management ([ac90947](https://github.com/FlameInTheDark/gochat/commit/ac909475869d2ba1c2bda5ac04abae87836a7dde))

# [1.2.0](https://github.com/FlameInTheDark/gochat/compare/v1.1.0...v1.2.0) (2025-09-20)


### Features

* guild invites ([80ece98](https://github.com/FlameInTheDark/gochat/commit/80ece98faed49d8f00dc1e670cbc32e7e791ed0d))

# [1.1.0](https://github.com/FlameInTheDark/gochat/compare/v1.0.1...v1.1.0) (2025-09-18)


### Bug Fixes

* chart debug ([4c2a860](https://github.com/FlameInTheDark/gochat/commit/4c2a8605152849d45581b944cd3b9816f43ac2cc))
* Guild route validation fixes ([69baf87](https://github.com/FlameInTheDark/gochat/commit/69baf87b80d6fbdf185d16ca983d71ea7b11e574))
* installer list style improvement ([477445d](https://github.com/FlameInTheDark/gochat/commit/477445d50106763c0bd0941d7741c0f3babcbe0c))
* installer rework ([81a8d72](https://github.com/FlameInTheDark/gochat/commit/81a8d729b6b668a3d4ff6d6ed8993404060c03fe))
* **installer:** set image pull policy to Always for api, ui, and ws ([ec38189](https://github.com/FlameInTheDark/gochat/commit/ec38189bb5ec204d1a541a4b677fc0e594ff7690))


### Features

* Added generation and js, go clients for the API. Fixed query and docs, added config files comments for easier navigation ([b830ca6](https://github.com/FlameInTheDark/gochat/commit/b830ca62b0c3465c36f28b8cc8370f031748d34e))
* Added member roles subroute to the guild route, bumped a Golang version ([73c3cd3](https://github.com/FlameInTheDark/gochat/commit/73c3cd3f35590b8157e5aeb97bb4eae70a919f89))
* added search endpoint (so far returns ids instead of messages), some minor refactoring ([6e5f74d](https://github.com/FlameInTheDark/gochat/commit/6e5f74dc2d630f8fd876aaf9cff7f54960feabd9))
* **api-deployment:** add S3 environment variables for MinIO configuration ([a47370a](https://github.com/FlameInTheDark/gochat/commit/a47370ac51f4c596caa78d05d4dc3431d2ffea3c))
* **api:** Migrated cold data from ScyllaDB to PostgreSQL ([5a216e2](https://github.com/FlameInTheDark/gochat/commit/5a216e264fc3694a9f80af2aebef5a40f9f4192b))
* **api:** Migration of cold data to PostgreSQL with Citus ([ed9f950](https://github.com/FlameInTheDark/gochat/commit/ed9f9509a69038316a8367f683c1c21452bd74fe))
* **api:** PostgreSQL migration files ([c6bb212](https://github.com/FlameInTheDark/gochat/commit/c6bb212251b61bd715cbe6f4a87b05c39790189f))
* **api:** request validation and some fixes ([50414da](https://github.com/FlameInTheDark/gochat/commit/50414daddf606e3f809f8d890360133125b5c52b))
* **api:** some refactoring, optimizations and idempotency ([ea9d0da](https://github.com/FlameInTheDark/gochat/commit/ea9d0da080e68af848770c2fbbcb0b76e86fa93c))
* changed refresh token route and update/reorder channels routesi ([9fb8070](https://github.com/FlameInTheDark/gochat/commit/9fb80708e8b71b155c766c93ed109b3808da152c))
* client docs and methods ([e465a2d](https://github.com/FlameInTheDark/gochat/commit/e465a2d49db374887ba25e0e28a94f2ccef68939))
* Improved authentication with access and refresh tokens ([c52a90d](https://github.com/FlameInTheDark/gochat/commit/c52a90dc9831ae2c97d4224ac065dd6001d450da))
* **ingress:** add MinIO Console ingress configuration ([afdd6c1](https://github.com/FlameInTheDark/gochat/commit/afdd6c1d793e132959fa6aa3a67b27d541077157))
* **installer:** enhance MinIO configuration with dedicated API credentials and console ingress options ([db21cd9](https://github.com/FlameInTheDark/gochat/commit/db21cd986f3c2353ffb0a92cc7fa310005f39b7b))
* JWT refresh, config file examples, Resend email provider, SMTP changes ([00db646](https://github.com/FlameInTheDark/gochat/commit/00db646688402426357e5a5699a959c8b9f19207))
* new auth service and password recovery ([3c35deb](https://github.com/FlameInTheDark/gochat/commit/3c35deb0d28a9e6734e4720a9b409a4b0ea5b986))
* removed helm (will add in feature), mailer rework in progress, makefile update for the easier deployment of dev environment, readme update for better information about the project ([af9e20b](https://github.com/FlameInTheDark/gochat/commit/af9e20b49a3fc0cac532116b89acf3fb3863c64f))
* search update ([0d87987](https://github.com/FlameInTheDark/gochat/commit/0d87987a4b55116143f9f7fc5ec3b3d927302fe2))
* **search:** a little optimization of the search query builder ([f12f1b0](https://github.com/FlameInTheDark/gochat/commit/f12f1b008c91ee5a49a55c0b53d7aa155a22c9a0))
* **search:** added OpenSearch and indexer service ([d060e20](https://github.com/FlameInTheDark/gochat/commit/d060e2055ac90ff6cf83e48b14d4b06e23085da5))
* **search:** updated search indexing for messages ([6c5104c](https://github.com/FlameInTheDark/gochat/commit/6c5104c64d58dd258f6742fea988f84dc7531e95))
* updated opensearch library and fixed document update on the indexer side ([b0359f5](https://github.com/FlameInTheDark/gochat/commit/b0359f544d68b08ff03f4957881ede5a98d729f7))
* updated token generation and refresh, updated smtp and fixed some issues with permissions check ([4a6a4c0](https://github.com/FlameInTheDark/gochat/commit/4a6a4c01a599e0be9bc30f96cc56119f901d6f05))

## [1.0.1](https://github.com/FlameInTheDark/gochat/compare/v1.0.0...v1.0.1) (2025-04-08)


### Bug Fixes

* **api:** fixed s3 event webhook panic ([6f87c86](https://github.com/FlameInTheDark/gochat/commit/6f87c86beec258a4a27c31a3adc12d9c9b6d082f))
* **helm:** fixed helm chart, added deployment with ingress ([924d840](https://github.com/FlameInTheDark/gochat/commit/924d8406d277671fed562c70b544f6181fe15e57))
* **installer:** added ability to install version from the dev branch ([ae7aa1f](https://github.com/FlameInTheDark/gochat/commit/ae7aa1faab0fcc551f1613815173cba0f3995862))
* **installer:** changed ws annotations ([cfa053f](https://github.com/FlameInTheDark/gochat/commit/cfa053f86e9023077fcb730ba16920df20754207))
* **installer:** changed ws annotations ([bca62cb](https://github.com/FlameInTheDark/gochat/commit/bca62cb394089bce4314ec90ac71c3c7635ffb2f))
* **installer:** deployment fixes ([2f792f3](https://github.com/FlameInTheDark/gochat/commit/2f792f3314c3fb597c657e89d943bbb9ee6ed837))
* **installer:** fixed ingress and database migration ([3f69e02](https://github.com/FlameInTheDark/gochat/commit/3f69e027f4d7989017caff7e14482af3bfff79da))
* **installer:** fixed installer context selection and server-snippets in values.yaml ([2e5fdb5](https://github.com/FlameInTheDark/gochat/commit/2e5fdb5a077a2f729b55d16f81ca98e535d9976d))
* **installer:** fixed websocket upgrade for ingress ([3b1980c](https://github.com/FlameInTheDark/gochat/commit/3b1980c519a01d233259a8c65836bfb91f666bac))
* **installer:** fixed ws routing issue ([75454be](https://github.com/FlameInTheDark/gochat/commit/75454bec6aab9680c0aafb3f85bda6fb48a03647))
* **installer:** fixed ws routing issue ([1a65ddc](https://github.com/FlameInTheDark/gochat/commit/1a65ddccdc1404109896a45d08623c054dd0fbf1))
* **installer:** values fix ([2a37bf8](https://github.com/FlameInTheDark/gochat/commit/2a37bf87cf3806b820344688922c1129777c2c5c))
* new deployment ([6352446](https://github.com/FlameInTheDark/gochat/commit/6352446211b97b0ea003d804df9ccc9423ccbf52))

# 1.0.0 (2025-04-05)


### Bug Fixes

* pipeline ([8c30c67](https://github.com/FlameInTheDark/gochat/commit/8c30c6739a5fe812dc97d7a4ba48545a281040b1))
* updated message deletion ([ef6f6dd](https://github.com/FlameInTheDark/gochat/commit/ef6f6ddf1deebc759609c4c02bf9a66f7775b612))
* **ws:** channel subscription ([40a1602](https://github.com/FlameInTheDark/gochat/commit/40a160227decc839e5f2783a281bdcd99ae7f9b9))
* **ws:** fixed dropped connection issues and heartbeat window +2 seconds ([2ea8fdd](https://github.com/FlameInTheDark/gochat/commit/2ea8fdd16c3cf6c9a0d70747d14f1c4e878d2918))
* **ws:** subscription handler fix ([7379b1c](https://github.com/FlameInTheDark/gochat/commit/7379b1c4ea04415818fcd5fe7be775a7cba9d17e))


### Features

* added channel threads, changed snowflake generation ([96fa73f](https://github.com/FlameInTheDark/gochat/commit/96fa73f4f3d04830ef408dda48a77a6d288d16a2))
* additional db fields ([e53ca8e](https://github.com/FlameInTheDark/gochat/commit/e53ca8e43a13eec81ac4f5c2ee51943163173232))
* api methods, websocket server improvement, logging and monitoring ([febaae4](https://github.com/FlameInTheDark/gochat/commit/febaae4c6c586a998daea76119402904ea5ba663))
* autoinstaller ([3c38b4e](https://github.com/FlameInTheDark/gochat/commit/3c38b4e2f120f3c3e2b6fe0a9ea4f104468cfded))
* improved error handling, added user routes ([774dab2](https://github.com/FlameInTheDark/gochat/commit/774dab2d00ca91eb929ff94e526e5daa3eaf05ce))
* initial implementation of the base structure of API ([378f0ef](https://github.com/FlameInTheDark/gochat/commit/378f0ef2dcc0699915f66c14c8ef052b1d678c7f))
* message bucketing and message update ([395b19b](https://github.com/FlameInTheDark/gochat/commit/395b19b41d2a3d7da7d327f4910330fc48f71533))
* migrations ([fd3b01f](https://github.com/FlameInTheDark/gochat/commit/fd3b01f4b2e815527e91c7b20920700f9fdc218a))
* too many additions to explain them all ([4bfc9cb](https://github.com/FlameInTheDark/gochat/commit/4bfc9cb0495190f6fffc8576eb59f60a2f73e39f))
* too many additions to explain them all ([a3c5230](https://github.com/FlameInTheDark/gochat/commit/a3c523088e244dcf0d352104b46585508d4c2926))
