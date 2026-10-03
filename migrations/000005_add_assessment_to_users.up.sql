ALTER TABLE users 
ADD COLUMN has_completed_assessment BOOLEAN DEFAULT FALSE,
ADD COLUMN assessment_data JSONB;
