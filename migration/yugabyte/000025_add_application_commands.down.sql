DROP INDEX IF EXISTS application_command_interactions_command_idx;
DROP INDEX IF EXISTS application_command_interactions_expiry_idx;
DROP INDEX IF EXISTS application_command_interactions_token_uidx;
DROP TABLE IF EXISTS application_command_interactions;

DROP INDEX IF EXISTS application_commands_application_idx;
DROP INDEX IF EXISTS application_commands_guild_idx;
DROP INDEX IF EXISTS application_commands_scope_name_uidx;
DROP TABLE IF EXISTS application_commands;
