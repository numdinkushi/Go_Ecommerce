-- Fix foreign key constraint violation for order_items
-- This script cleans up orphaned order_items before recreating the constraint

-- Step 1: Drop the existing foreign key constraint if it exists
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_orders_items;

-- Step 2: Find and delete orphaned order_items (items referencing non-existent orders)
DELETE FROM order_items 
WHERE order_id NOT IN (SELECT id FROM orders);

-- Step 3: Verify cleanup
-- Check for any remaining orphaned records
SELECT COUNT(*) as orphaned_items
FROM order_items oi
LEFT JOIN orders o ON oi.order_id = o.id
WHERE o.id IS NULL;

-- The constraint will be recreated automatically by GORM's AutoMigrate
-- If you need to add it manually, uncomment the next line:
-- ALTER TABLE order_items ADD CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;

