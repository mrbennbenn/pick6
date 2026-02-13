-- Fix swapped choice_a and choice_b for Joe vs Bahaa and Dee vs Monique matches
-- The images show fighters in specific left/right positions that don't match the database

-- Swap Joe Brooks and Bahaa Kabil
UPDATE questions 
SET choice_a = 'Bahaa Kabil', choice_b = 'Joe Brooks'
WHERE question_id = 'question_39aJ1eE9ihQ3hH9kmOfKdCSueFP';

-- Swap Dee Begley and Monique Ettiene
UPDATE questions 
SET choice_a = 'Monique Ettiene', choice_b = 'Dee Begley'
WHERE question_id = 'question_39aJ1lcWNS9J0c9WODFGxAgzHXR';

-- Update existing user responses to maintain vote accuracy
-- Flip all 'a' votes to 'b' and all 'b' votes to 'a' for these two questions
UPDATE responses 
SET choice = CASE 
    WHEN choice = 'a' THEN 'b'
    WHEN choice = 'b' THEN 'a'
END
WHERE question_id IN (
    'question_39aJ1eE9ihQ3hH9kmOfKdCSueFP',
    'question_39aJ1lcWNS9J0c9WODFGxAgzHXR'
);
