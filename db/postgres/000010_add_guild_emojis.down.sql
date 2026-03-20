-- +migrate StatementBegin
DO $$
BEGIN
    EXECUTE 'DROP TABLE IF EXISTS emoji_lookup';
    EXECUTE 'DROP TABLE IF EXISTS guild_emojis';
END
$$;
-- +migrate StatementEnd
