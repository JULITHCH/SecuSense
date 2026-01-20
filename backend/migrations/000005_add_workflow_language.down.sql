-- Remove language column from course_workflow_sessions
ALTER TABLE course_workflow_sessions
DROP COLUMN IF EXISTS language;
