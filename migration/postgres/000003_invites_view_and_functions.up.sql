-- Active (non-expired) invites view
CREATE OR REPLACE VIEW active_guild_invites AS
SELECT *
FROM guild_invites
WHERE expires_at > now();

-- Lookup function: resolve code, purge if expired, return if valid
CREATE OR REPLACE FUNCTION fetch_guild_invite(p_code varchar)
    RETURNS TABLE (
        invite_code varchar(8),
        invite_id   bigint,
        guild_id    bigint,
        author_id   bigint,
        created_at  timestamptz,
        expires_at  timestamptz
        )
    LANGUAGE plpgsql
        SECURITY DEFINER
SET search_path = public, pg_temp
    AS $$
DECLARE
  v_invite_id bigint;
  v_guild_id  bigint;
  v_deleted   integer;
BEGIN
  -- Resolve code (single-shard on mapper)
  SELECT ic.invite_id, ic.guild_id
    INTO v_invite_id, v_guild_id
    FROM guild_invite_codes ic
   WHERE ic.invite_code = p_code;

  IF NOT FOUND THEN
    RETURN; -- no such code
  END IF;

  -- TTL delete on the guild shard
  DELETE FROM guild_invites gi
   WHERE gi.guild_id  = v_guild_id
     AND gi.invite_id = v_invite_id
     AND gi.expires_at <= now();
  GET DIAGNOSTICS v_deleted = ROW_COUNT;

  IF v_deleted > 0 THEN
    -- Remove orphaned mapper row
    DELETE FROM guild_invite_codes
     WHERE invite_code = p_code;
    RETURN; -- expired and removed
  END IF;

  -- Return valid invite
  RETURN QUERY
  SELECT p_code, gi.invite_id, gi.guild_id, gi.author_id, gi.created_at, gi.expires_at
    FROM guild_invites gi
   WHERE gi.guild_id  = v_guild_id
     AND gi.invite_id = v_invite_id
     AND gi.expires_at > now();
END;
$$;

-- Helper to delete by (guild_id, invite_id) and also clean up the mapper (use this instead of triggers)
CREATE OR REPLACE FUNCTION delete_guild_invite(p_guild_id bigint, p_invite_id bigint)
    RETURNS void
    LANGUAGE plpgsql
        SECURITY DEFINER
SET search_path = public, pg_temp
    AS $$
BEGIN
  DELETE FROM guild_invites
   WHERE guild_id = p_guild_id
     AND invite_id = p_invite_id;

  DELETE FROM guild_invite_codes
   WHERE guild_id = p_guild_id
     AND invite_id = p_invite_id;
END;
$$;

-- Optional: grant execute to your app role
-- GRANT EXECUTE ON FUNCTION fetch_guild_invite(varchar) TO your_app_role;
-- GRANT EXECUTE ON FUNCTION delete_guild_invite(bigint, bigint) TO your_app_role;
