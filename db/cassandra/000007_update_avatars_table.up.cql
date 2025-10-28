ALTER TABLE gochat.avatars DROP object;
ALTER TABLE gochat.avatars ADD url text;
ALTER TABLE gochat.avatars ADD content_type text;
ALTER TABLE gochat.avatars ADD width bigint;
ALTER TABLE gochat.avatars ADD height bigint;
ALTER TABLE gochat.avatars ADD filesize bigint;
ALTER TABLE gochat.avatars ADD done boolean;

