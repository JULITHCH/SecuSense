-- Remove video generation status from courses
DROP INDEX IF EXISTS idx_courses_video_status;
ALTER TABLE courses DROP COLUMN IF EXISTS video_error;
ALTER TABLE courses DROP COLUMN IF EXISTS video_status;
