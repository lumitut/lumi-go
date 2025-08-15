-- Database seed script for development environment
-- This script populates the database with sample data for testing

-- Ensure we're using the correct database
\c lumigo;

-- Insert additional test users
INSERT INTO users (email, username, full_name, role, status, email_verified, password_hash)
VALUES
    ('john.doe@example.com', 'johndoe', 'John Doe', 'user', 'active', true,
     '$2a$10$YKqH9J.7rZFGqK.3XvVwUOqGqK8mKBhV8vJ2fYQYLmKv6BzXvXXXX'),
    ('jane.smith@example.com', 'janesmith', 'Jane Smith', 'user', 'active', true,
     '$2a$10$YKqH9J.7rZFGqK.3XvVwUOqGqK8mKBhV8vJ2fYQYLmKv6BzXvXXXX'),
    ('bob.wilson@example.com', 'bobwilson', 'Bob Wilson', 'user', 'pending', false,
     '$2a$10$YKqH9J.7rZFGqK.3XvVwUOqGqK8mKBhV8vJ2fYQYLmKv6BzXvXXXX'),
    ('alice.johnson@example.com', 'alicej', 'Alice Johnson', 'admin', 'active', true,
     '$2a$10$YKqH9J.7rZFGqK.3XvVwUOqGqK8mKBhV8vJ2fYQYLmKv6BzXvXXXX'),
    ('test.inactive@example.com', 'inactive', 'Inactive User', 'user', 'inactive', true,
     '$2a$10$YKqH9J.7rZFGqK.3XvVwUOqGqK8mKBhV8vJ2fYQYLmKv6BzXvXXXX')
ON CONFLICT (email) DO NOTHING;

-- Insert sample API keys
INSERT INTO api_keys (user_id, name, key_hash, scopes, metadata)
SELECT
    u.id,
    'Development API Key',
    '$2a$10$' || encode(gen_random_bytes(32), 'hex'),
    ARRAY['read', 'write'],
    '{"environment": "development", "created_by": "seed_script"}'::jsonb
FROM users u
WHERE u.email = 'john.doe@example.com'
ON CONFLICT DO NOTHING;

-- Insert sample feature flags for testing
INSERT INTO feature_flags (name, description, enabled, rules, metadata)
VALUES
    ('beta_features', 'Enable beta features for testing', true,
     '{"percentage": 100, "user_ids": [], "environments": ["dev", "staging"]}'::jsonb,
     '{"category": "experimental"}'::jsonb),
    ('maintenance_mode', 'Put application in maintenance mode', false,
     '{"message": "System maintenance in progress", "allowed_ips": ["127.0.0.1"]}'::jsonb,
     '{"category": "operational"}'::jsonb),
    ('new_onboarding', 'New user onboarding flow', true,
     '{"percentage": 50, "cohort": "new_users"}'::jsonb,
     '{"category": "ux"}'::jsonb),
    ('enhanced_security', 'Enhanced security features', true,
     '{"require_2fa": false, "session_timeout": 3600}'::jsonb,
     '{"category": "security"}'::jsonb)
ON CONFLICT (name) DO UPDATE SET
    enabled = EXCLUDED.enabled,
    rules = EXCLUDED.rules,
    metadata = EXCLUDED.metadata;

-- Insert sample audit logs
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
SELECT
    u.id,
    'user.login',
    'user',
    u.id,
    '{"method": "password", "success": true}'::jsonb,
    '192.168.1.100',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'
FROM users u
WHERE u.email IN ('john.doe@example.com', 'jane.smith@example.com');

-- Insert sample sessions (for testing session management)
INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
SELECT
    u.id,
    'dev_refresh_token_' || encode(gen_random_bytes(16), 'hex'),
    'Mozilla/5.0 (Development Environment)',
    '127.0.0.1',
    NOW() + INTERVAL '7 days'
FROM users u
WHERE u.status = 'active' AND u.email_verified = true
ON CONFLICT DO NOTHING;

-- Create some test data patterns for load testing
DO $$
DECLARE
    i INTEGER;
    user_id UUID;
BEGIN
    -- Get a test user ID
    SELECT id INTO user_id FROM users WHERE email = 'john.doe@example.com';

    -- Create multiple audit log entries for testing pagination
    FOR i IN 1..50 LOOP
        INSERT INTO audit_logs (user_id, action, entity_type, metadata, ip_address, created_at)
        VALUES (
            user_id,
            CASE
                WHEN i % 5 = 0 THEN 'user.logout'
                WHEN i % 3 = 0 THEN 'user.update'
                ELSE 'user.login'
            END,
            'user',
            ('{"iteration": ' || i || ', "test": true}')::jsonb,
            '192.168.1.' || (i % 255),
            NOW() - (i || ' hours')::INTERVAL
        );
    END LOOP;
END $$;

-- Output summary
DO $$
DECLARE
    user_count INTEGER;
    session_count INTEGER;
    flag_count INTEGER;
    audit_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO user_count FROM users;
    SELECT COUNT(*) INTO session_count FROM sessions WHERE revoked_at IS NULL;
    SELECT COUNT(*) INTO flag_count FROM feature_flags;
    SELECT COUNT(*) INTO audit_count FROM audit_logs;

    RAISE NOTICE '';
    RAISE NOTICE '=================================';
    RAISE NOTICE 'Database Seeding Complete';
    RAISE NOTICE '=================================';
    RAISE NOTICE 'Users created: %', user_count;
    RAISE NOTICE 'Active sessions: %', session_count;
    RAISE NOTICE 'Feature flags: %', flag_count;
    RAISE NOTICE 'Audit log entries: %', audit_count;
    RAISE NOTICE '';
    RAISE NOTICE 'Test Credentials:';
    RAISE NOTICE '  Admin: admin@lumitut.com / admin123';
    RAISE NOTICE '  User: user@lumitut.com / user123';
    RAISE NOTICE '  Test: john.doe@example.com / password';
    RAISE NOTICE '=================================';
END $$;
