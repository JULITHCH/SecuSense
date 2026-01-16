-- Add video generation status to courses
ALTER TABLE courses ADD COLUMN IF NOT EXISTS video_status VARCHAR(50) DEFAULT NULL;
ALTER TABLE courses ADD COLUMN IF NOT EXISTS video_error TEXT DEFAULT NULL;

-- Add index for finding courses with pending video generation
CREATE INDEX IF NOT EXISTS idx_courses_video_status ON courses(video_status) WHERE video_status IS NOT NULL;
