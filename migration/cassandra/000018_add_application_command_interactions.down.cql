DROP TABLE IF EXISTS gochat.application_command_interactions;

ALTER TABLE gochat.messages DROP interaction_user_id;
ALTER TABLE gochat.messages DROP interaction_command_name;
ALTER TABLE gochat.messages DROP interaction_command_id;
ALTER TABLE gochat.messages DROP interaction_application_id;
ALTER TABLE gochat.messages DROP interaction_id;
