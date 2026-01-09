-- SQL query to check recent transactions and their SessionIDs
-- Run this to see what SessionIDs are stored in your database

-- Get all recent transactions with their SessionIDs (stored as transaction_id)
SELECT 
    id,
    user_id,
    order_id,
    transaction_id as session_id_stored,  -- This is the SessionID stored in DB
    payment_id,
    status,
    created_at
FROM transactions
ORDER BY created_at DESC
LIMIT 10;

-- Get the most recent transaction for a specific order
-- Replace ORDER_ID with your order ID
-- SELECT transaction_id as session_id_stored
-- FROM transactions
-- WHERE order_id = ORDER_ID
-- ORDER BY created_at DESC
-- LIMIT 1;

