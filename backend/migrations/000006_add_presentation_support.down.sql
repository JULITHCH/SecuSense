-- Drop presentations table
DROP TABLE IF EXISTS lesson_presentations;

-- Remove columns from lesson_scripts
ALTER TABLE lesson_scripts
DROP COLUMN IF EXISTS output_type,
DROP COLUMN IF EXISTS presentation_status;
