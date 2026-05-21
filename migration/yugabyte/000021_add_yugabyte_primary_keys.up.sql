ALTER TABLE dm_channels
    ADD CONSTRAINT dm_channels_pkey PRIMARY KEY USING INDEX idx_unique_dm_channel_by_channel_user;

ALTER TABLE channel_roles_permissions
    ADD CONSTRAINT channel_roles_permissions_pkey PRIMARY KEY USING INDEX idx_unique_channel_role_perm;

ALTER TABLE members
    ADD CONSTRAINT members_pkey PRIMARY KEY USING INDEX idx_unique_member;

ALTER TABLE channel_user_permissions
    ADD CONSTRAINT channel_user_permissions_pkey PRIMARY KEY USING INDEX idx_unique_channel_user_perm;

ALTER TABLE friends
    ADD CONSTRAINT friends_pkey PRIMARY KEY USING INDEX idx_unique_friend;

ALTER TABLE user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY USING INDEX idx_unique_user_role;

ALTER TABLE thread_members
    ADD CONSTRAINT thread_members_pkey PRIMARY KEY USING INDEX idx_thread_members_thread_id_user_id;
