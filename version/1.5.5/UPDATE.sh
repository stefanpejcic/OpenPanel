#!/bin/bash

mysql <<EOF

USE panel;

-- Check if 'container' column exists, and only add it if it doesn't
SET @col_exists := (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE table_schema = DATABASE()
    AND table_name = 'sites'
    AND column_name = 'container'
);

-- Prepare and execute the ALTER TABLE only if the column is missing
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE sites ADD COLUMN container VARCHAR(255) DEFAULT NULL;',
  'SELECT "Column already exists, skipping ALTER TABLE."');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Now do the update
UPDATE sites
SET container = ports,
    ports = NULL;

EOF
