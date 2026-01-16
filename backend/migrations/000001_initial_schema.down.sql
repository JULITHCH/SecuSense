-- Drop triggers
DROP TRIGGER IF EXISTS update_ai_jobs_updated_at ON ai_generation_jobs;
DROP TRIGGER IF EXISTS update_tests_updated_at ON tests;
DROP TRIGGER IF EXISTS update_enrollments_updated_at ON enrollments;
DROP TRIGGER IF EXISTS update_courses_updated_at ON courses;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS ai_generation_jobs;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS user_answers;
DROP TABLE IF EXISTS test_attempts;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS tests;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS course_content;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
