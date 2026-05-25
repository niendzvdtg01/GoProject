-- Migration to add role column to users table
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'member' AFTER password_hash;