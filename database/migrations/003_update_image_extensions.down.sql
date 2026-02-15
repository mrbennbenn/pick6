-- Revert image filenames back to .png extensions
-- This is the rollback for the image optimization migration

UPDATE questions 
SET image_filename = REPLACE(image_filename, '.jpg', '.png')
WHERE image_filename LIKE '%.jpg';
