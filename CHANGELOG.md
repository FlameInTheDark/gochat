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
