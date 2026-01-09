# Stripe Webhook Setup Guide

## The Problem

Stripe **cannot** send webhooks to `localhost` URLs. If you're running your server locally, Stripe has no way to reach `http://localhost:8080/webhooks/payment/stripe` because:
- Stripe's servers are on the internet
- Your localhost is only accessible on your machine
- Stripe needs a **publicly accessible URL** to send webhooks

## Solution: Use Stripe CLI or ngrok (for Local Development)

### Option 1: Stripe CLI (Recommended for Local Testing)

**Stripe CLI is the easiest way to forward webhooks to localhost.**

#### Step 1: Install Stripe CLI

```bash
# macOS
brew install stripe/stripe-cli/stripe

# Or download from https://stripe.com/docs/stripe-cli
```

#### Step 2: Authenticate Stripe CLI

```bash
stripe login
```

#### Step 3: Start Your Server

```bash
# In your project directory
make run
# or
go run main.go
```

Your server should be running on `http://localhost:8080` (or whatever port you configured).

#### Step 4: Forward Webhooks with Stripe CLI

```bash
# Replace 8080 with your actual server port
stripe listen --forward-to localhost:8080/webhooks/payment/stripe
```

You'll see output like:
```
> Ready! Your webhook signing secret is whsec_xxxxxxxxxxxxx (^C to quit)
```

**Copy the webhook signing secret** (starts with `whsec_...`)

#### Step 5: Add Webhook Secret to .env

```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
```

**Benefits of Stripe CLI:**
- ✅ Automatically forwards webhooks from Stripe Dashboard
- ✅ Works with both test and live mode
- ✅ Provides webhook signing secret automatically
- ✅ No need to configure webhook URL in Stripe Dashboard
- ✅ Better for development workflow

### Option 2: Use ngrok (Alternative)

#### Step 1: Install ngrok

```bash
# macOS
brew install ngrok

# Or download from https://ngrok.com/download
```

#### Step 2: Start Your Server

```bash
# In your project directory
make run
# or
go run main.go
```

Your server should be running on `http://localhost:8080` (or whatever port you configured).

#### Step 3: Expose Your Server with ngrok

```bash
# Replace 8080 with your actual server port
ngrok http 8080
```

You'll see output like:
```
Forwarding  https://abc123.ngrok.io -> http://localhost:8080
```

**Copy the HTTPS URL** (e.g., `https://abc123.ngrok.io`)

### Step 4: Configure Webhook in Stripe Dashboard (Only for ngrok)

**Note:** If using Stripe CLI, skip this step. Stripe CLI automatically forwards webhooks.

1. Go to [Stripe Dashboard](https://dashboard.stripe.com)
2. Navigate to **Developers** → **Webhooks**
3. Click **Add endpoint**
4. Enter your webhook URL:
   ```
   https://abc123.ngrok.io/webhooks/payment/stripe
   ```
   (Replace `abc123.ngrok.io` with your actual ngrok URL)
5. **Select events to listen to:**
   - ✅ `checkout.session.completed`
   - ✅ `payment_intent.succeeded`
   - ✅ `charge.succeeded`
   - ✅ `charge.updated` (optional, for status updates)
6. Click **Add endpoint**

### Step 5: Get Your Webhook Secret

**If using Stripe CLI:** The webhook secret was already shown in Step 4 above. Copy it from the terminal output.

**If using ngrok:**
1. After creating the endpoint, click on it
2. Click **Reveal** next to "Signing secret"
3. Copy the secret (starts with `whsec_...`)

4. Add it to your `.env` file:
   ```bash
   STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
   ```

### Step 6: Restart Your Server

```bash
# Stop and restart your server to load the new env var
make run
```

## Testing

### Test 1: Verify Webhook Endpoint is Reachable

**If using ngrok:**
```bash
# Test the health check endpoint
curl https://abc123.ngrok.io/webhooks/payment/test
```

**If using Stripe CLI:**
```bash
# Test the health check endpoint
curl http://localhost:8080/webhooks/payment/test
```

Should return:
```json
{"message": "Webhook endpoint is reachable"}
```

### Test 2: Make a Real Payment

1. Create a checkout session through your app
2. Complete the payment in Stripe's test mode
3. **Stripe will automatically send webhooks:**
   - If using Stripe CLI: Webhooks are automatically forwarded
   - If using ngrok: Webhooks are sent to your ngrok URL
4. Check your server logs - you should see webhook events being received

### Test 3: Trigger Test Webhook (Stripe CLI Only)

If using Stripe CLI, you can trigger test webhooks directly:

```bash
stripe trigger checkout.session.completed
```

This sends a test webhook to your local server without making a real payment.

### Test 4: Check Webhook Logs in Stripe Dashboard

1. Go to **Developers** → **Webhooks** → Your endpoint
2. Click on **Events** tab
3. You should see webhook events being sent automatically
4. Green checkmarks = successful delivery
5. Red X = failed delivery (check your server logs)

**Note:** With Stripe CLI, webhook deliveries are shown in the CLI terminal, not in Dashboard.

## Important Notes

### Stripe CLI vs ngrok

**Stripe CLI (Recommended):**
- ✅ No URL configuration needed in Stripe Dashboard
- ✅ Automatically handles webhook forwarding
- ✅ Works seamlessly with test and live mode
- ✅ Better developer experience

**ngrok:**
- ⚠️ Free URLs change every time you restart ngrok
- ⚠️ Requires manual webhook URL configuration in Stripe Dashboard
- ⚠️ Must update webhook URL if ngrok restarts
- ✅ Paid plan available for static URLs

### Production Setup

For production, you don't need ngrok:
- Deploy your server to a public URL (e.g., `https://api.yourapp.com`)
- Configure webhook endpoint: `https://api.yourapp.com/webhooks/payment/stripe`
- Use the same steps above, but with your production URL

### Webhook Events

The following events are automatically sent by Stripe when payments complete:

1. **checkout.session.completed** - When user completes checkout
   - Contains: `SessionID`, `PaymentID`, `Metadata` (with `order_id`)

2. **payment_intent.succeeded** - When payment succeeds
   - Contains: `PaymentID`, `Metadata` (if configured)

3. **charge.succeeded** - When charge succeeds
   - Contains: `PaymentID` (from PaymentIntent)

**You don't need to manually trigger these** - Stripe sends them automatically!

## Troubleshooting

### Webhooks Not Being Sent

1. ✅ Check ngrok is running: `curl https://your-ngrok-url.ngrok.io/webhooks/payment/test`
2. ✅ Verify webhook URL in Stripe Dashboard matches your ngrok URL
3. ✅ Check that events are enabled in Stripe Dashboard
4. ✅ Verify `STRIPE_WEBHOOK_SECRET` is set in your `.env`
5. ✅ Check server logs for incoming webhook requests

### Webhooks Failing

1. Check server logs for errors
2. Verify webhook signature validation is working
3. Check that transaction matching logic is working
4. Look at Stripe Dashboard → Webhooks → Events for error details

### Test Webhooks vs Real Webhooks

- **Test webhooks** from Stripe Dashboard use dummy IDs that won't match your transactions
- **Real webhooks** from actual payments will have real PaymentIDs that match your database
- The code now handles test webhooks gracefully (returns success without processing)

## Quick Reference

**Webhook Endpoint:** `/webhooks/payment/stripe`

**Required Environment Variables:**
```bash
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

**Required Events:**
- `checkout.session.completed`
- `payment_intent.succeeded`
- `charge.succeeded`

**Webhook URL Format:**
```
https://your-domain.com/webhooks/payment/stripe
```

