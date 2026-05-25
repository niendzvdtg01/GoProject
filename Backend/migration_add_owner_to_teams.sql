-- Migration to add owner_id column to teams table
ALTER TABLE teams ADD COLUMN owner_id VARCHAR(36) NOT NULL AFTER team_name;