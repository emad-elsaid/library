-- up
ALTER TABLE books
  ADD COLUMN page_read integer NOT NULL DEFAULT 0;

-- down
ALTER TABLE books
  DROP COLUMN page_read integer NOT NULL DEFAULT 0;
