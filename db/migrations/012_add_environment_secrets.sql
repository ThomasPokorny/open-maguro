-- +goose Up
ALTER TABLE skills ADD COLUMN environment_secrets TEXT;

-- +goose Down
ALTER TABLE skills DROP COLUMN IF EXISTS environment_secrets;
