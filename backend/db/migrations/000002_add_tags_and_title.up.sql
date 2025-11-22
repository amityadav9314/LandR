ALTER TABLE materials ADD COLUMN IF NOT EXISTS title VARCHAR(255);
UPDATE materials SET title = 'Untitled Material' WHERE title IS NULL;
ALTER TABLE materials ALTER COLUMN title SET DEFAULT 'Untitled Material';

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE TABLE IF NOT EXISTS material_tags (
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (material_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);
