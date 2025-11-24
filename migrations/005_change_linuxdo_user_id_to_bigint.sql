-- Change linuxdo_user_id from VARCHAR to BIGINT
ALTER TABLE users ALTER COLUMN linuxdo_user_id TYPE BIGINT USING linuxdo_user_id::BIGINT;
