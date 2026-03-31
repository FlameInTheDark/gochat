ALTER TABLE gochat.avatars DROP done;
ALTER TABLE gochat.avatars DROP filesize;
ALTER TABLE gochat.avatars DROP height;
ALTER TABLE gochat.avatars DROP width;
ALTER TABLE gochat.avatars DROP content_type;
ALTER TABLE gochat.avatars DROP url;
ALTER TABLE gochat.avatars ADD object text;
