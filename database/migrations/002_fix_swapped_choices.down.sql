-- Rollback: Revert the swapped choices back to original state

-- Revert Joe Brooks and Bahaa Kabil to original positions
UPDATE questions 
SET choice_a = 'Joe Brooks', choice_b = 'Bahaa Kabil'
WHERE question_id = 'question_39aJ1eE9ihQ3hH9kmOfKdCSueFP';

-- Revert Dee Begley and Monique Ettiene to original positions
UPDATE questions 
SET choice_a = 'Dee Begley', choice_b = 'Monique Ettiene'
WHERE question_id = 'question_39aJ1lcWNS9J0c9WODFGxAgzHXR';

-- Revert user responses back to original state
UPDATE responses 
SET choice = CASE 
    WHEN choice = 'a' THEN 'b'
    WHEN choice = 'b' THEN 'a'
END
WHERE question_id IN (
    'question_39aJ1eE9ihQ3hH9kmOfKdCSueFP',
    'question_39aJ1lcWNS9J0c9WODFGxAgzHXR'
);
