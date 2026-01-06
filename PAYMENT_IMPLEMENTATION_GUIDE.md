# Payment Implementation Guide

## Table of Contents
1. [Current Implementation State](#current-implementation-state)
2. [Payment Flow Options](#payment-flow-options)
3. [Recommended Implementation Strategy](#recommended-implementation-strategy)
4. [Phase-by-Phase Implementation Plan](#phase-by-phase-implementation-plan)
5. [Payment Gateway Integration Details](#payment-gateway-integration-details)
6. [Database Schema Considerations](#database-schema-considerations)
7. [Security Best Practices](#security-best-practices)
8. [Testing Procedures](#testing-procedures)
9. [Error Handling & Edge Cases](#error-handling--edge-cases)
10. [Migration Path from Current State](#migration-path-from-current-state)

---

## Current Implementation State

### What's Currently Implemented

**Location:** `backend/internal/service/userService.go` - `CreateOrder` method

**Current Code:**
```go
// find success payment reference status
// TODO: Implement payment verification logic
paymentID := "PAY12345"
transactionID := "12345"
_ = paymentID
_ = transactionID
```

**Current Behavior:**
- Orders are created with hardcoded payment IDs
- `paymentID`: Always `"PAY12345"`
- `transactionID`: Always `"12345"`
- Order status: Always `"pending"`
- No actual payment processing occurs
- No payment verification happens

**Existing Infrastructure:**
- Flutterwave integration exists for bank verification (`BankService`)
- Flutterwave client is initialized in `server.go`
- Payment gateway credentials are in config (`FlutterwaveSecretKey`, `FlutterwaveClientID`, `FlutterwaveEncryptionKey`)

**Problems with Current Implementation:**
1. ❌ Orders created without payment verification
2. ❌ No way to track real payment transactions
3. ❌ No integration with payment gateway
4. ❌ Risk of creating orders for unpaid items
5. ❌ Cannot process refunds or handle payment disputes
6. ❌ No payment status tracking
7. ❌ Cannot support multiple payment methods

---

## Payment Flow Options

### Option 1: Pre-Payment Flow (Recommended for Most E-commerce)

**Description:** Payment is processed and verified BEFORE order creation.

**Flow Diagram:**
```
User → Add to Cart → Checkout → Initialize Payment → 
Payment Gateway → User Pays → Webhook/Callback → 
Verify Payment → Create Order → Confirm Order
```

**Step-by-Step Process:**

1. **User Initiates Checkout**
   - User clicks "Checkout" or "Place Order"
   - System calculates total amount
   - System validates cart items are still available

2. **Initialize Payment**
   - Create payment intent/transaction with payment gateway
   - Payment gateway returns:
     - `paymentID` (unique payment reference)
     - `transactionID` (transaction identifier)
     - Payment URL or payment token (for redirect)

3. **User Completes Payment**
   - User redirected to payment gateway (or in-app payment)
   - User enters payment details
   - Payment gateway processes payment

4. **Payment Gateway Callback**
   - Payment gateway sends webhook/callback to your server
   - Callback includes:
     - Payment status (success/failed)
     - Transaction ID
     - Payment ID
     - Amount paid
     - Payment method

5. **Verify Payment & Create Order**
   - Server verifies webhook signature (security)
   - Check payment status is "successful"
   - Verify amount matches order total
   - Create order with real payment IDs
   - Set order status to "paid" or "processing"

6. **Order Confirmation**
   - Send confirmation email to user
   - Clear cart
   - Return order details to user

**Benefits:**
- ✅ Payment verified before order creation
- ✅ Real transaction tracking
- ✅ Supports refunds and disputes
- ✅ Industry standard approach
- ✅ Better fraud prevention
- ✅ Automatic payment status updates

**Drawbacks:**
- ⚠️ More complex implementation
- ⚠️ Requires webhook infrastructure
- ⚠️ User must complete payment before order exists

**Best For:**
- Online stores with card payments
- High-volume e-commerce
- International transactions
- Automated fulfillment systems

---

### Option 2: Post-Payment Flow (Cash on Delivery / Manual Verification)

**Description:** Order is created first, payment verified later.

**Flow Diagram:**
```
User → Add to Cart → Create Order (Pending) → 
Generate Payment Reference → User Pays Later → 
Admin Verifies → Update Order Status
```

**Step-by-Step Process:**

1. **Create Order Immediately**
   - User clicks "Place Order"
   - System creates order with status "pending"
   - Generate internal payment reference ID
   - Store payment method (COD, bank transfer, etc.)

2. **Generate Payment Reference**
   - Create unique payment reference (UUID or sequential)
   - Store in `paymentID` field
   - Generate transaction reference in `transactionID`
   - Display to user for payment

3. **User Pays Later**
   - For COD: Payment on delivery
   - For bank transfer: User transfers money using reference
   - For manual payment: User pays via other means

4. **Admin/System Verifies Payment**
   - Admin checks payment manually
   - Or system checks bank account for transfer
   - Or payment gateway callback confirms payment

5. **Update Order**
   - Update `paymentID` and `transactionID` with real values
   - Change order status to "paid"
   - Send confirmation to user

**Benefits:**
- ✅ Simpler initial implementation
- ✅ Works for COD (Cash on Delivery)
- ✅ Good for manual payment verification
- ✅ Faster MVP development
- ✅ Works in regions with limited payment gateways

**Drawbacks:**
- ⚠️ Manual verification required (for some methods)
- ⚠️ Risk of unpaid orders
- ⚠️ Not scalable for high volume
- ⚠️ Requires admin intervention
- ⚠️ Slower order fulfillment

**Best For:**
- Cash on Delivery (COD) businesses
- Local businesses with manual verification
- MVP/testing phase
- Regions with limited payment gateway support
- B2B transactions with invoice-based payments

---

### Option 3: Hybrid Approach (Best of Both Worlds)

**Description:** Support multiple payment methods with different flows.

**Flow Diagram:**
```
User → Checkout → Select Payment Method →
├─ Card/Online → Pre-Payment Flow
├─ COD → Post-Payment Flow
└─ Bank Transfer → Post-Payment Flow
```

**Step-by-Step Process:**

1. **User Selects Payment Method**
   - Display available payment options
   - User selects: Card, COD, Bank Transfer, etc.

2. **Route Based on Payment Method**
   - **Card/Online Payment:**
     - Initialize payment gateway
     - Process payment immediately
     - Create order after payment confirmation
   - **COD:**
     - Create order immediately
     - Generate payment reference
     - Mark as "pending_payment"
     - Verify on delivery
   - **Bank Transfer:**
     - Create order immediately
     - Generate unique reference number
     - User transfers money
     - System/Admin verifies transfer
     - Update order status

3. **Unified Order Management**
   - All orders stored in same table
   - Different statuses based on payment method
   - Unified tracking and reporting

**Benefits:**
- ✅ Flexible payment methods
- ✅ Supports both automated and manual flows
- ✅ Can migrate gradually
- ✅ Better user experience (choice)
- ✅ Works in diverse markets

**Drawbacks:**
- ⚠️ More complex to implement
- ⚠️ Requires handling multiple flows
- ⚠️ More testing scenarios

**Best For:**
- Multi-region businesses
- Businesses offering multiple payment options
- Gradual migration from manual to automated
- Businesses serving diverse customer base

---

## Recommended Implementation Strategy

### Overall Recommendation: **Hybrid Approach with Phased Implementation**

**Why Hybrid?**
- Provides flexibility for different payment methods
- Allows gradual migration from current state
- Supports both automated and manual verification
- Better user experience with payment options

**Why Phased?**
- Reduces risk by implementing incrementally
- Allows testing at each phase
- Can launch MVP faster
- Easier to debug and fix issues

---

## Phase-by-Phase Implementation Plan

### Phase 1: Immediate Improvements (MVP - Week 1-2)

**Goal:** Replace hardcoded values with proper internal references

**Tasks:**

1. **Generate Unique Payment References**
   - Replace hardcoded `paymentID` with UUID generation
   - Replace hardcoded `transactionID` with sequential ID or UUID
   - Format: `PAY-{UUID}` or `PAY-{timestamp}-{random}`

2. **Add Payment Method Field**
   - Add `payment_method` to Order domain
   - Options: "cod", "card", "bank_transfer", "pending"
   - Default to "pending" for now

3. **Add Payment Status Field**
   - Add `payment_status` to Order domain
   - Options: "pending", "processing", "paid", "failed", "refunded"
   - Separate from order status

4. **Update CreateOrder Logic**
   - Generate unique payment references
   - Set payment_method to "pending"
   - Set payment_status to "pending"
   - Keep order status as "pending"

5. **Create Payment Verification Endpoint**
   - `POST /orders/:id/verify-payment`
   - Allows manual payment verification
   - Updates payment_status and payment IDs
   - Admin-only or automated

**Deliverables:**
- ✅ Unique payment references generated
- ✅ Payment method tracking
- ✅ Payment status tracking
- ✅ Manual verification endpoint

**Testing:**
- Verify unique payment IDs are generated
- Test payment verification endpoint
- Ensure no duplicate payment IDs

---

### Phase 2: Payment Gateway Integration (Weeks 3-6)

**Goal:** Integrate real payment gateway for card/online payments

**Tasks:**

1. **Extend Flutterwave Integration**
   - Currently only used for bank verification
   - Add payment processing capabilities
   - Implement payment initialization
   - Implement webhook handling

2. **Create Payment Service**
   - New service: `PaymentService`
   - Methods:
     - `InitializePayment(amount, currency, userID)`
     - `VerifyPayment(paymentID)`
     - `ProcessWebhook(webhookData)`
     - `RefundPayment(paymentID, amount)`

3. **Create Checkout Endpoint**
   - `POST /checkout`
   - Validates cart
   - Calculates total
   - Initializes payment with gateway
   - Returns payment URL/token

4. **Create Payment Webhook Handler**
   - `POST /webhooks/payment`
   - Verifies webhook signature
   - Processes payment status
   - Creates order on successful payment
   - Updates order on payment failure

5. **Update CreateOrder Flow**
   - Accept payment details as parameter
   - Verify payment before creating order
   - Use real payment IDs from gateway
   - Set appropriate statuses

6. **Payment Status Polling (Optional)**
   - Background job to check pending payments
   - Updates orders if payment completed
   - Handles cases where webhook fails

**Deliverables:**
- ✅ Payment gateway integration
- ✅ Checkout endpoint
- ✅ Webhook handler
- ✅ Payment verification
- ✅ Order creation after payment

**Testing:**
- Test payment initialization
- Test successful payment flow
- Test failed payment flow
- Test webhook handling
- Test payment verification

---

### Phase 3: Enhanced Features (Weeks 7-10)

**Goal:** Add advanced payment features and optimizations

**Tasks:**

1. **Payment Retry Logic**
   - Allow users to retry failed payments
   - Store payment attempts
   - Limit retry attempts
   - Notify user of retry options

2. **Refund Processing**
   - Implement refund endpoint
   - Process refunds through gateway
   - Update order status
   - Send refund notifications

3. **Payment Status Tracking**
   - Real-time payment status updates
   - Payment history per order
   - Payment event logging

4. **Order Cancellation**
   - Cancel order if payment fails
   - Cancel order if payment timeout
   - Restore cart items on cancellation
   - Notify user of cancellation

5. **Email Notifications**
   - Payment initiated
   - Payment successful
   - Payment failed
   - Payment pending
   - Refund processed

6. **Payment Analytics**
   - Track payment success rates
   - Monitor payment failures
   - Payment method preferences
   - Revenue tracking

**Deliverables:**
- ✅ Payment retry functionality
- ✅ Refund processing
- ✅ Enhanced notifications
- ✅ Order cancellation
- ✅ Payment analytics

**Testing:**
- Test retry logic
- Test refund processing
- Test cancellation flow
- Test notifications
- Verify analytics

---

## Payment Gateway Integration Details

### Flutterwave Integration (Recommended)

**Why Flutterwave?**
- Already partially integrated (bank verification)
- Good for African markets
- Supports multiple payment methods
- Good documentation and support

**Required Credentials:**
- `FLUTTERWAVE_SECRET_KEY` - Already in config
- `FLUTTERWAVE_PUBLIC_KEY` - May need to add
- `FLUTTERWAVE_ENCRYPTION_KEY` - Already in config

**Key Flutterwave Endpoints:**

1. **Initialize Payment**
   - Endpoint: `POST /v3/payments`
   - Creates payment link/transaction
   - Returns payment URL and transaction reference

2. **Verify Payment**
   - Endpoint: `GET /v3/transactions/{id}/verify`
   - Checks payment status
   - Returns transaction details

3. **Webhook**
   - Flutterwave sends POST to your webhook URL
   - Contains payment status and details
   - Must verify webhook signature

4. **Refund**
   - Endpoint: `POST /v3/refunds`
   - Processes refunds
   - Returns refund status

**Implementation Structure:**
```
pkg/external/flutterwave/
  ├── client.go (existing - extend)
  ├── payment.go (new - payment methods)
  ├── webhook.go (new - webhook handling)
  └── types.go (new - payment types)
```

---

### Alternative Payment Gateways

**Stripe:**
- Best for international markets
- Excellent documentation
- Strong fraud prevention
- Higher fees

**Paystack:**
- Popular in Nigeria
- Good for African markets
- Similar to Flutterwave
- Good developer experience

**PayPal:**
- Global reach
- User trust
- Higher fees
- More complex integration

---

## Database Schema Considerations

### Current Order Schema
```go
type Order struct {
    ID             uint
    UserId         uint
    Status         string      // "pending", "processing", etc.
    Amount         float64
    TransactionId  string      // Currently hardcoded
    OrderRefNumber uint
    PaymentId      string      // Currently hardcoded
    Items          []OrderItem
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### Recommended Additions

**Option 1: Extend Order Domain**
```go
type Order struct {
    // ... existing fields ...
    PaymentMethod    string      // "card", "cod", "bank_transfer"
    PaymentStatus    string      // "pending", "paid", "failed", "refunded"
    PaymentVerifiedAt *time.Time
    PaymentGateway   string      // "flutterwave", "stripe", "manual"
}
```

**Option 2: Separate Payment Table (Recommended for Complex Systems)**
```go
type Payment struct {
    ID              uint
    OrderID         uint
    PaymentMethod   string
    PaymentStatus   string
    PaymentID       string      // Gateway payment ID
    TransactionID   string      // Gateway transaction ID
    Amount          float64
    Currency        string
    Gateway         string
    GatewayResponse json.RawMessage
    VerifiedAt      *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Benefits of Separate Payment Table:**
- ✅ Multiple payments per order (retries, partial payments)
- ✅ Payment history tracking
- ✅ Better audit trail
- ✅ Easier refund processing
- ✅ More flexible for complex scenarios

---

## Security Best Practices

### 1. Webhook Security

**Verify Webhook Signatures:**
- Never trust webhook data without verification
- Payment gateways provide signature in headers
- Verify signature using secret key
- Reject webhooks with invalid signatures

**Example:**
```
X-Flutterwave-Signature: <signature>
```

### 2. Payment ID Security

**Never Trust Client-Provided Payment IDs:**
- Always verify payment IDs server-side
- Check payment belongs to correct user
- Verify payment amount matches order
- Use server-side API calls to verify

### 3. Idempotency

**Prevent Duplicate Processing:**
- Use idempotency keys for payment operations
- Store processed webhook IDs
- Check if webhook already processed
- Prevent duplicate order creation

### 4. Amount Verification

**Always Verify Amounts:**
- Never trust client-provided amounts
- Recalculate from cart items server-side
- Verify payment amount matches order total
- Reject mismatched amounts

### 5. User Verification

**Verify Payment Ownership:**
- Ensure payment belongs to authenticated user
- Check user ID matches payment user ID
- Prevent cross-user payment access
- Validate user permissions

### 6. HTTPS Only

**Secure Communication:**
- All payment endpoints must use HTTPS
- Never send payment data over HTTP
- Use secure webhook URLs
- Encrypt sensitive payment data

---

## Testing Procedures

### Unit Testing

**Test Payment Reference Generation:**
- Verify unique payment IDs
- Test ID format
- Test ID collision prevention

**Test Payment Verification:**
- Test successful verification
- Test failed verification
- Test invalid payment IDs
- Test amount verification

### Integration Testing

**Test Payment Gateway Integration:**
- Test payment initialization
- Test payment verification
- Test webhook processing
- Test refund processing

**Test Order Creation Flow:**
- Test order creation after payment
- Test order creation failure scenarios
- Test cart clearing
- Test email notifications

### End-to-End Testing

**Complete Payment Flow:**
1. Add items to cart
2. Initiate checkout
3. Process payment
4. Verify webhook received
5. Verify order created
6. Verify cart cleared
7. Verify email sent

**Failure Scenarios:**
1. Payment fails → Order not created
2. Webhook timeout → Polling updates order
3. Payment amount mismatch → Reject payment
4. Invalid webhook signature → Reject webhook

### Manual Testing Checklist

- [ ] Payment initialization works
- [ ] Payment gateway redirect works
- [ ] Successful payment creates order
- [ ] Failed payment doesn't create order
- [ ] Webhook processing works
- [ ] Payment verification works
- [ ] Refund processing works
- [ ] Email notifications sent
- [ ] Cart cleared after order
- [ ] Payment IDs stored correctly
- [ ] Order status updated correctly
- [ ] Error handling works

---

## Error Handling & Edge Cases

### Common Error Scenarios

**1. Payment Gateway Unavailable**
- **Handling:** Queue payment for retry
- **User Experience:** Show "Payment processing, please try again"
- **Recovery:** Retry payment initialization

**2. Webhook Timeout**
- **Handling:** Implement polling fallback
- **User Experience:** Order status shows "processing"
- **Recovery:** Poll payment gateway for status

**3. Payment Amount Mismatch**
- **Handling:** Reject payment, don't create order
- **User Experience:** Show error message
- **Recovery:** User must retry with correct amount

**4. Duplicate Webhook**
- **Handling:** Check if webhook already processed
- **User Experience:** No impact (idempotent)
- **Recovery:** Ignore duplicate webhook

**5. Payment Succeeded but Order Creation Failed**
- **Handling:** Create order from payment data
- **User Experience:** Order appears in account
- **Recovery:** Manual reconciliation if needed

**6. Order Created but Payment Failed**
- **Handling:** Cancel order, restore cart
- **User Experience:** Notification of cancellation
- **Recovery:** User can retry checkout

### Edge Cases to Handle

- Payment partially processed
- Network timeout during payment
- User closes browser during payment
- Multiple payment attempts
- Payment refund after order fulfillment
- Currency conversion issues
- Payment gateway rate limiting

---

## Migration Path from Current State

### Step 1: Prepare Database (No Breaking Changes)

1. Add new fields to Order table:
   - `payment_method` (nullable, default NULL)
   - `payment_status` (nullable, default "pending")
   - `payment_verified_at` (nullable)

2. Migrate existing orders:
   - Set `payment_method` to "pending" for existing orders
   - Set `payment_status` to "pending" for existing orders
   - Keep existing `paymentID` and `transactionID` values

### Step 2: Update Code (Backward Compatible)

1. Update `CreateOrder` to generate unique IDs
2. Add payment method parameter (optional)
3. Set defaults for new fields
4. Maintain backward compatibility

### Step 3: Deploy Phase 1

1. Deploy code with new payment reference generation
2. Monitor for issues
3. Verify new orders have unique payment IDs
4. Test payment verification endpoint

### Step 4: Add Payment Gateway (New Features)

1. Implement payment gateway integration
2. Add checkout endpoint
3. Add webhook handler
4. Test thoroughly in staging

### Step 5: Deploy Phase 2

1. Deploy payment gateway integration
2. Enable for new orders
3. Monitor payment success rates
4. Gradually migrate users

### Step 6: Enhancements

1. Add retry logic
2. Add refund processing
3. Add analytics
4. Optimize based on data

---

## Implementation Checklist

### Phase 1: MVP Improvements
- [ ] Generate unique payment references
- [ ] Add payment_method field to Order
- [ ] Add payment_status field to Order
- [ ] Update CreateOrder to use new fields
- [ ] Create payment verification endpoint
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Update API documentation

### Phase 2: Payment Gateway
- [ ] Extend Flutterwave client
- [ ] Create PaymentService
- [ ] Implement payment initialization
- [ ] Implement webhook handler
- [ ] Create checkout endpoint
- [ ] Update CreateOrder flow
- [ ] Add webhook signature verification
- [ ] Write tests
- [ ] Update documentation

### Phase 3: Enhanced Features
- [ ] Implement payment retry
- [ ] Implement refund processing
- [ ] Add email notifications
- [ ] Add order cancellation
- [ ] Add payment analytics
- [ ] Optimize performance
- [ ] Write tests
- [ ] Update documentation

---

## Monitoring & Maintenance

### Key Metrics to Track

1. **Payment Success Rate**
   - Percentage of successful payments
   - Track by payment method
   - Identify failure patterns

2. **Payment Processing Time**
   - Time from initiation to completion
   - Identify slow payments
   - Optimize bottlenecks

3. **Webhook Processing**
   - Webhook delivery success rate
   - Webhook processing time
   - Failed webhook recovery

4. **Order Creation Rate**
   - Orders created per payment
   - Failed order creation after payment
   - Reconciliation needs

5. **Refund Rate**
   - Percentage of refunds
   - Refund processing time
   - Refund success rate

### Logging Requirements

- Log all payment initializations
- Log all webhook receipts
- Log payment verifications
- Log order creations
- Log payment failures
- Log refunds

### Alerting

- Alert on high payment failure rate
- Alert on webhook processing failures
- Alert on payment gateway downtime
- Alert on order creation failures

---

## Conclusion

This guide provides a comprehensive roadmap for implementing proper payment processing in your e-commerce system. The phased approach allows for gradual implementation while maintaining system stability.

**Key Takeaways:**
1. Start with Phase 1 to replace hardcoded values
2. Integrate payment gateway in Phase 2
3. Add enhancements in Phase 3
4. Always prioritize security
5. Test thoroughly at each phase
6. Monitor and optimize continuously

**Next Steps:**
1. Review this guide with your team
2. Choose payment gateway (Flutterwave recommended)
3. Plan Phase 1 implementation
4. Set up development environment
5. Begin Phase 1 implementation

For questions or clarifications, refer to the payment gateway documentation or consult with your development team.

