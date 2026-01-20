-- Add output type and presentation status to lesson_scripts
ALTER TABLE lesson_scripts
ADD COLUMN IF NOT EXISTS output_type VARCHAR(20) NOT NULL DEFAULT 'video',
ADD COLUMN IF NOT EXISTS presentation_status VARCHAR(20);

-- Create presentations table
CREATE TABLE IF NOT EXISTS lesson_presentations (
    id UUID PRIMARY KEY,
    lesson_id UUID NOT NULL REFERENCES lesson_scripts(id) ON DELETE CASCADE,
    slides JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create index for fast lookup by lesson_id
CREATE INDEX IF NOT EXISTS idx_lesson_presentations_lesson_id ON lesson_presentations(lesson_id);
