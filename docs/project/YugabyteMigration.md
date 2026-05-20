# YugabyteDB Migration

[<- Documentation](README.md)

This runbook moves GoChat relational data from the legacy PostgreSQL/Citus local stack to YugabyteDB YSQL. Data loss is a release blocker: do not remove, prune, or overwrite the source Citus containers, volumes, or backups until export, import, verification, and smoke tests have all passed.

## Local Docker Compose Cutover

1. Freeze write-capable services before exporting relational data:
   ```bash
   docker compose stop api auth attachments search ws webhook
   ```

2. Start or keep the legacy source Citus stack available under the explicit profile:
   ```bash
   make citus_up
   ```

3. Take a source backup before running Voyager:
   ```bash
   mkdir -p .data/db-dumps
   docker compose exec -T citus-master pg_dump -U postgres -d gochat -Fc > .data/db-dumps/gochat-citus-before-yugabyte.dump
   ```

4. Assess the source with YugabyteDB Voyager:
   ```bash
   make voyager_assess
   ```

   The Make targets run the official Voyager container on the Compose network by default. To use a host-installed Voyager binary instead, override both the command and export directory:
   ```bash
   make voyager_assess VOYAGER=yb-voyager VOYAGER_EXPORT_DIR=.data/voyager/gochat CITUS_HOST=127.0.0.1
   ```

5. Start YugabyteDB and create the `gochat` database:
   ```bash
   make yugabyte_up
   ```

   The local init service creates new YugabyteDB databases with `WITH COLOCATION = false` by default. Core relational metadata stays non-colocated and sharded so hot tables can split across tablets as the cluster grows. Existing databases are not converted in place; if `yb_is_database_colocated()` returns the wrong layout for the target, create a new database and migrate into it. Do not drop or recreate a database that may contain application data.

6. Apply the Yugabyte-compatible schema migrations:
   ```bash
   make migrate_yugabyte
   ```

   If the host already has another database bound to `127.0.0.1:5433`, run the migration image from the Compose network instead:
   ```bash
   docker run --rm --network gochat_default \
     -e MIGRATION_SCOPE=yugabyte \
     -e YUGABYTE_ADDRESS="postgres://yugabyte:yugabyte@yugabyte:5433/gochat?sslmode=disable" \
     gochat-migrations:local
   ```

7. Export and import application data with Voyager. The Make target excludes `schema_migrations` because the target migration runner owns that table.
   ```bash
   make voyager_export
   make voyager_import
   make voyager_status
   ```

8. Verify source and target tables:
   ```bash
   make yugabyte_verify
   ```

   Expected result starts with:
   ```text
   verification=OK schema=public checksum=true
   ```

9. Restart services against YugabyteDB:
   ```bash
   docker compose up -d api auth attachments search ws webhook
   ```

10. Run API/auth/search/ws smoke tests that create, read, update, and delete relational records. Keep the Citus profile and backup until the smoke run is accepted.

## Tooling

- Active local schema: `migration/yugabyte`
- Legacy Citus schema: `migration/postgres`
- Default full migration scope: YugabyteDB YSQL + ScyllaDB
- Legacy scope: `MIGRATION_SCOPE=citus|postgres|pg`
- Voyager default: Dockerized `software.yugabyte.com/yugabytedb/yb-voyager` on `gochat_default`
- YugabyteDB database colocation: disabled by default with `YUGABYTE_COLOCATION=false`
- Verification command:
  ```bash
  make yugabyte_verify
  ```

`gctools yugabyte verify` compares all base tables in the selected schema by row count and deterministic row checksums. Use `--skip-checksum` only when a table is too large for checksum aggregation and a separate validation method has been recorded.

## High-load YugabyteDB layout

The production posture is a non-colocated YugabyteDB YSQL database for users, guilds, channels, members, roles, invites, permissions, and auth metadata. YugabyteDB colocation is useful for tiny reference datasets or small isolated databases, but colocated tables share one parent tablet and colocated tablets do not tablet-split. That makes database-wide colocation the wrong default for Discord-scale guild and member traffic.

For guild startup reads, keep the source tables sharded and optimize the access pattern instead of pinning all guild data to one tablet:

- Keep ScyllaDB as the high-volume message, read-state, attachment, and event timeline store.
- Keep guild-scoped YSQL tables keyed and indexed by `guild_id` where the query is guild-local, for example members, roles, guild channels, invites, emoji, and discovery stats.
- Add materialized/read-model tables or cache entries for the client "guild ready" bundle when a single request needs guild data plus channels and role metadata.
- Use `SPLIT INTO` or cluster-level shard settings for hot tables and indexes after measuring table size, node count, and write rate; avoid creating excess tablets for small tables.
- Only colocate tiny, low-write reference tables after they are proven not to be on hot paths.

For an existing local database that was created colocated, use a side-by-side target instead of mutating it in place:

```bash
docker compose exec -T yugabyte /home/yugabyte/bin/ysqlsh \
  -h yugabyte -p 5433 -U yugabyte -d yugabyte \
  -c "CREATE DATABASE gochat_sharded WITH COLOCATION = false"

docker run --rm --network gochat_default \
  -e MIGRATION_SCOPE=yugabyte \
  -e YUGABYTE_ADDRESS="postgres://yugabyte:yugabyte@yugabyte:5433/gochat_sharded?sslmode=disable" \
  gochat-migrations:local
```

Then copy from the current source of truth into `gochat_sharded`, verify `gochat -> gochat_sharded` with `gctools yugabyte verify`, switch service DSNs, and keep both databases until smoke tests pass.

When copying between two YugabyteDB databases in the same local Compose stack, prefer a data-only dump from the current active database after write-capable services have been stopped:

```bash
docker run --rm --network gochat_default \
  -e PGPASSWORD=yugabyte \
  -v "./.data/db-dumps:/dumps" \
  postgres:15-alpine pg_dump \
  -h yugabyte -p 5433 -U yugabyte -d gochat \
  --data-only \
  --exclude-table=schema_migrations \
  --exclude-schema=ybvoyager_metadata \
  --no-owner --no-privileges \
  -f /dumps/gochat-yugabyte-to-sharded-data.sql

docker run --rm --network gochat_default \
  -e PGPASSWORD=yugabyte \
  -v "./.data/db-dumps:/dumps:ro" \
  postgres:15-alpine psql \
  -h yugabyte -p 5433 -U yugabyte -d gochat_sharded \
  -v ON_ERROR_STOP=1 \
  -f /dumps/gochat-yugabyte-to-sharded-data.sql
```

## Rollback

- If Voyager export fails, keep services stopped and investigate the source; no target data is trusted.
- If import or verification fails, keep the current source database as the source of truth, leave application configs pointed back at the previous DSN only after explicitly deciding to abort, and preserve `.data/db-dumps/gochat-citus-before-yugabyte.dump`.
- If services fail after restart, stop write-capable services again, switch configs back to Citus only if verification showed target divergence, and record which check failed.
- Do not run `docker compose down -v`, remove Citus containers, or delete source backups as part of rollback.

## Kubernetes Migration Shape

Use an offline maintenance window for the first cluster cutover:

1. Announce maintenance and stop all deployment replicas that can write relational data.
2. Take a managed PostgreSQL/Citus backup and verify it can be restored.
3. Deploy YugabyteDB with the official Helm chart and create the target database/user with `COLOCATION = false` for the main GoChat relational database.
4. Apply `migration/yugabyte` to the target.
5. Run Voyager assessment, export, and import from a migration job with network access to both databases.
6. Run `gctools yugabyte verify` from inside the cluster or a trusted bastion.
7. Update Kubernetes secrets/config maps from the Citus DSN to the YugabyteDB YSQL DSN.
8. Roll out services gradually and smoke test relational workflows before ending maintenance.
9. Keep source Citus read-only and backed up until post-cutover monitoring is accepted.

## References

- [YugabyteDB Docker quick start](https://docs.yugabyte.com/stable/quick-start/docker/)
- [YugabyteDB Voyager](https://docs.yugabyte.com/stable/yugabyte-voyager/)
- [Voyager offline migration steps](https://docs.yugabyte.com/stable/yugabyte-voyager/migrate/migrate-steps/)
- [YugabyteDB Kubernetes Helm deployment](https://docs.yugabyte.com/stable/deploy/kubernetes/single-zone/oss/helm-chart/)
- [YugabyteDB colocation](https://docs.yugabyte.com/stable/additional-features/colocation/)
- [YSQL data modeling and performance](https://docs.yugabyte.com/stable/develop/best-practices-develop/data-modeling-perf/)
- [YugabyteDB tablet splitting](https://docs.yugabyte.com/stable/architecture/docdb-sharding/tablet-splitting/)
