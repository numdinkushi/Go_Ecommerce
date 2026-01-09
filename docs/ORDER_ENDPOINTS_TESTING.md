# Order Endpoints Testing Guide

## Prerequisites
- Server should be running (default: `http://localhost:8080` or check your server config)
- You'll need a valid authentication token for all order endpoints
- Products must exist in the database (or create them first via seller endpoints)

---

## Step 1: Register a User (or Login)

### Option A: Register New User
**Endpoint:** `POST /register`

**Headers:**
```
Content-Type: application/json
```

**Payload:**
```json
{
  "email": "testuser@example.com",
  "password": "password123",
  "first_name": "Test",
  "last_name": "User",
  "phone": "+1234567890"
}
```

**Expected Response:**
```json
{
  "message": "User registered successfully",
  "user": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Save the `token` from the response for subsequent requests.**

---

### Option B: Login Existing User
**Endpoint:** `POST /login`

**Headers:**
```
Content-Type: application/json
```

**Payload:**
```json
{
  "email": "testuser@example.com",
  "password": "password123"
}
```

**Expected Response:**
```json
{
  "message": "login",
  "user": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Save the `token` from the response.**

---

## Step 2: Add Items to Cart

**Note:** Orders are created from cart items, so you need items in your cart first.

**Endpoint:** `POST /cart`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <your_token_here>
```

**Payload:**
```json
{
  "product_id": 1,
  "quantity": 2
}
```

**Expected Response:**
```json
{
  "message": "Item added to cart successfully",
  "cart": {
    "id": 1,
    "product_id": 1,
    "quantity": 2,
    "price": 29.99,
    ...
  }
}
```

**Repeat this step to add multiple items to cart if needed.**

**Verify Cart:**
**Endpoint:** `GET /cart`

**Headers:**
```
Authorization: Bearer <your_token_here>
```

**No payload required.**

---

## Step 3: Create an Order

**Endpoint:** `POST /orders`

**Headers:**
```
Content-Type: application/json
Authorization: Bearer <your_token_here>
```

**Payload:**
```json
{}
```

**Note:** The `CreateOrder` endpoint doesn't require a body - it uses all items from the user's cart.

**Expected Response:**
```json
{
  "message": "order created successfully",
  "order_id": 1
}
```

**Save the `order_id` for testing the get order by ID endpoint.**

**Note:** After creating an order, the cart will be automatically cleared.

---

## Step 4: Get All Orders

**Endpoint:** `GET /orders`

**Headers:**
```
Authorization: Bearer <your_token_here>
```

**No query parameters or payload required.**

**Expected Response:**
```json
{
  "message": "Orders retrieved successfully",
  "orders": [
    {
      "id": 1,
      "user_id": 1,
      "status": "pending",
      "amount": 59.98,
      "transaction_id": "12345",
      "order_ref_number": 12345678,
      "payment_id": "PAY12345",
      "items": [
        {
          "id": 1,
          "product_id": 1,
          "name": "Product Name",
          "image_url": "https://...",
          "seller_id": 1,
          "price": 29.99,
          "quantity": 2,
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z"
        }
      ],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

---

## Step 5: Get Order by ID

**Endpoint:** `GET /orders/:id`

**Replace `:id` with the actual order ID from Step 3.**

**Headers:**
```
Authorization: Bearer <your_token_here>
```

**No payload required.**

**Example:** `GET /orders/1`

**Expected Response:**
```json
{
  "message": "Order retrieved successfully",
  "order": {
    "id": 1,
    "user_id": 1,
    "status": "pending",
    "amount": 59.98,
    "transaction_id": "12345",
    "order_ref_number": 12345678,
    "payment_id": "PAY12345",
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "name": "Product Name",
        "image_url": "https://...",
        "seller_id": 1,
        "price": 29.99,
        "quantity": 2,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Error Cases:**
- **404 Not Found:** If order doesn't exist or doesn't belong to the user
- **401 Unauthorized:** If token is missing or invalid

---

## Testing with cURL

### 1. Register/Login
```bash
# Register
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User",
    "phone": "+1234567890"
  }'

# Or Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "password123"
  }'
```

### 2. Add to Cart
```bash
curl -X POST http://localhost:8080/cart \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

### 3. Create Order
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{}'
```

### 4. Get All Orders
```bash
curl -X GET http://localhost:8080/orders \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 5. Get Order by ID
```bash
curl -X GET http://localhost:8080/orders/1 \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## Testing with Postman/Insomnia

1. **Create a new collection** for Order Testing
2. **Set up environment variables:**
   - `base_url`: `http://localhost:8080`
   - `token`: (will be set after login)
3. **Create requests in order:**
   - Register/Login → Save token to environment
   - Add to Cart → Use token from environment
   - Create Order → Use token from environment
   - Get All Orders → Use token from environment
   - Get Order by ID → Use token from environment

---

## Important Notes

1. **Authentication Required:** All order endpoints require a valid JWT token in the `Authorization` header
2. **Cart Dependency:** You must have items in your cart before creating an order
3. **User Isolation:** Users can only see and access their own orders
4. **Cart Clearing:** After creating an order, the cart is automatically cleared
5. **Order Status:** New orders are created with status "pending"
6. **Order Reference:** Each order gets a unique `order_ref_number` generated automatically

---

## Error Scenarios to Test

1. **Create Order with Empty Cart:**
   - Expected: Error message "cart is empty"

2. **Get Order by ID - Not Found:**
   - Use a non-existent order ID
   - Expected: 404 with "Order not found"

3. **Get Order by ID - Wrong User:**
   - Create order with User A
   - Try to access with User B's token
   - Expected: 404 with "Order not found"

4. **Missing Authentication:**
   - Call endpoints without Authorization header
   - Expected: 401 Unauthorized

5. **Invalid Token:**
   - Use an expired or invalid token
   - Expected: 401 Unauthorized

