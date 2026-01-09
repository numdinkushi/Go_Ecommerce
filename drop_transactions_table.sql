-- SQL script to drop the transactions table
-- Run this in your PostgreSQL database

-- Option 1: Drop table with CASCADE (removes all dependent objects like foreign keys)
DROP TABLE IF EXISTS transactions CASCADE;

-- Option 2: Drop table only if it exists (safe version)
-- DROP TABLE IF EXISTS transactions;

-- Option 3: Drop table and all related constraints manually
-- First, drop foreign key constraints if they exist
-- ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_transaction;
-- ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_transaction;
-- Then drop the table
-- DROP TABLE IF EXISTS transactions;

-- Verify table is dropped
-- SELECT table_name 
-- FROM information_schema.tables 
-- WHERE table_schema = 'public' 
-- AND table_name = 'transactions';

