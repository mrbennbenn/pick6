-- Update image filenames to match optimized files
-- Changes .png extensions to .jpg for all matchup images that were optimized
-- This aligns the database with the actual filesystem after image optimization

UPDATE questions 
SET image_filename = REPLACE(image_filename, '.png', '.jpg')
WHERE image_filename LIKE '%.png';
