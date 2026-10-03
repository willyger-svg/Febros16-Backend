CREATE TABLE IF NOT EXISTS user_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_streak INT DEFAULT 0,
    activities_completed INT DEFAULT 0,
    triggers_identified INT DEFAULT 0,
    last_checkin_date DATE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_progress_user_id ON user_progress(user_id);
