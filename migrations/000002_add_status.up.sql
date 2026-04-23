ALTER TABLE tracks ADD COLUMN status VARCHAR(20) DEFAULT 'uploading';

UPDATE tracks SET status = 'ready' WHERE status IS NULL;

ALTER TABLE tracks ALTER COLUMN status SET NOT NULL;