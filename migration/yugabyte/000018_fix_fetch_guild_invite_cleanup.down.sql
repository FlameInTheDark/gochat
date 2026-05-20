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
  SELECT ic.invite_id, ic.guild_id
    INTO v_invite_id, v_guild_id
    FROM guild_invite_codes ic
   WHERE ic.invite_code = p_code;

  IF NOT FOUND THEN
    RETURN;
  END IF;

  DELETE FROM guild_invites gi
   WHERE gi.guild_id  = v_guild_id
     AND gi.invite_id = v_invite_id
     AND gi.expires_at <= now();
  GET DIAGNOSTICS v_deleted = ROW_COUNT;

  IF v_deleted > 0 THEN
    DELETE FROM guild_invite_codes
     WHERE invite_code = p_code;
    RETURN;
  END IF;

  RETURN QUERY
  SELECT p_code, gi.invite_id, gi.guild_id, gi.author_id, gi.created_at, gi.expires_at
    FROM guild_invites gi
   WHERE gi.guild_id  = v_guild_id
     AND gi.invite_id = v_invite_id
     AND gi.expires_at > now();
END;
$$;
