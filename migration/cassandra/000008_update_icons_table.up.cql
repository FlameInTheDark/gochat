ALTER TABLE gochat.icons DROP object;
ALTER TABLE gochat.icons ADD url text;
ALTER TABLE gochat.icons ADD content_type text;
ALTER TABLE gochat.icons ADD width bigint;
ALTER TABLE gochat.icons ADD height bigint;
ALTER TABLE gochat.icons ADD filesize bigint;
ALTER TABLE gochat.icons ADD done boolean;

