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
