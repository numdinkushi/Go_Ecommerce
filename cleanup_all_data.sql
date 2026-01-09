-- SQL script to safely delete all payment/order data
-- Deletes in correct order to avoid foreign key constraint violations

-- Step 1: Delete order_items first (they reference orders)
DELETE FROM order_items;

-- Step 2: Delete transactions (they reference orders)
DELETE FROM transactions;

-- Step 3: Delete payments (if they reference transactions)
DELETE FROM payments;

-- Step 4: Delete orders (now safe since nothing references them)
DELETE FROM orders;

-- Step 5: Verify cleanup
SELECT 
    (SELECT COUNT(*) FROM order_items) as order_items_count,
    (SELECT COUNT(*) FROM transactions) as transactions_count,
    (SELECT COUNT(*) FROM payments) as payments_count,
    (SELECT COUNT(*) FROM orders) as orders_count;

-- All counts should be 0 after running this script

