-- Add language column to course_workflow_sessions
ALTER TABLE course_workflow_sessions
ADD COLUMN IF NOT EXISTS language VARCHAR(10) NOT NULL DEFAULT 'en';
