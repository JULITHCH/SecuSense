-- Course Workflow Sessions (multi-agency course creation)
CREATE TABLE IF NOT EXISTS course_workflow_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    main_topic VARCHAR(500) NOT NULL,
    target_audience VARCHAR(255),
    difficulty_level VARCHAR(50),
    video_duration_min INTEGER DEFAULT 5,
    current_step VARCHAR(50) NOT NULL DEFAULT 'research',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Topic Suggestions (Step 1: Research Agency)
CREATE TABLE IF NOT EXISTS topic_suggestions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES course_workflow_sessions(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    is_custom BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Refined Topics (Step 2: Refinement Agency)
CREATE TABLE IF NOT EXISTS refined_topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES course_workflow_sessions(id) ON DELETE CASCADE,
    suggestion_id UUID NOT NULL REFERENCES topic_suggestions(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    learning_goals JSONB DEFAULT '[]',
    estimated_time_min INTEGER DEFAULT 10,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lesson Scripts (Step 3: Script Agency)
CREATE TABLE IF NOT EXISTS lesson_scripts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES course_workflow_sessions(id) ON DELETE CASCADE,
    topic_id UUID NOT NULL REFERENCES refined_topics(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    script TEXT NOT NULL,
    duration_min INTEGER DEFAULT 5,
    video_id VARCHAR(100),
    video_url TEXT,
    video_status VARCHAR(50),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_topic_suggestions_session ON topic_suggestions(session_id);
CREATE INDEX idx_topic_suggestions_status ON topic_suggestions(status);
CREATE INDEX idx_refined_topics_session ON refined_topics(session_id);
CREATE INDEX idx_lesson_scripts_session ON lesson_scripts(session_id);
CREATE INDEX idx_workflow_sessions_status ON course_workflow_sessions(status);

-- Trigger for updated_at
CREATE TRIGGER update_workflow_sessions_updated_at
    BEFORE UPDATE ON course_workflow_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
