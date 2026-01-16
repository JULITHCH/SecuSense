-- Add default admin user
-- Email: admin@secusense.local
-- Password: admin123 (bcrypt hash with cost 10)
-- IMPORTANT: Change this password immediately after first login!

INSERT INTO users (id, email, password_hash, first_name, last_name, role)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin@secusense.local',
    '$2b$10$YovPLIURpuqX3yVz2Spv4OmwfkiULjZpIaleJddHCd36pb.QHjIL2',
    'System',
    'Administrator',
    'admin'
) ON CONFLICT (email) DO NOTHING;
