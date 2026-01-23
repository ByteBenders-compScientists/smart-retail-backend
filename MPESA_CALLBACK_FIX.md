# M-Pesa Callback Issue - FIXED ✅

## Problem Summary
M-Pesa payments were stuck in "pending" status indefinitely because the callback URL configured in the environment variables did not match the actual route registered in the application.

## Root Cause
**URL Path Mismatch:**
- **Environment Variable**: `MPESA_CALLBACK_URL=https://drinx-retailers-bd.onrender.com/api/v1/mpesa/callback`
- **Actual Route** (before fix): `/api/v1/payments/mpesa/callback` ❌
- **Fixed Route**: `/api/v1/mpesa/callback` ✅

When M-Pesa tried to send payment confirmations to `/api/v1/mpesa/callback`, it received a 404 error because the route was registered at `/api/v1/payments/mpesa/callback`. This caused all payment callbacks to fail silently, leaving payments perpetually in "pending" status.

## Changes Made

### 1. Fixed Route Configuration
**File**: `internals/api/routes.go`

**Before:**
```go
api.POST("/payments/mpesa/callback", controllers.MpesaCallback)
```

**After:**
```go
api.POST("/mpesa/callback", controllers.MpesaCallback)
```

This now correctly matches the URL: `https://drinx-retailers-bd.onrender.com/api/v1/mpesa/callback`

### 2. Enhanced Logging
**File**: `internals/controllers/payments.go`

Added detailed logging to help diagnose callback issues:
- Log request metadata (IP address, user agent, method, path)
- Log raw request body when JSON parsing fails
- Better visibility into callback reception and processing

**Benefits:**
- Easier troubleshooting of callback issues
- Can identify if callbacks are being received but failing to parse
- Can verify the source of callback requests

## Testing

### Manual Testing with curl

Test the callback endpoint directly:

```bash
# Test with the provided script
./test_mpesa_callback.sh

# Or test manually
curl -X POST https://drinx-retailers-bd.onrender.com/api/v1/mpesa/callback \
  -H "Content-Type: application/json" \
  -d '{
    "Body": {
      "stkCallback": {
        "MerchantRequestID": "test-123",
        "CheckoutRequestID": "ws_CO_test",
        "ResultCode": 0,
        "ResultDesc": "Success"
      }
    }
  }'
```

Expected response: `{"ResultCode":0,"ResultDesc":"Accepted"}` with HTTP 200

### End-to-End Testing

1. **Initiate a payment** via the frontend or API
2. **Complete the payment** on your phone when prompted
3. **Wait 30 seconds** for M-Pesa to send the callback
4. **Check payment status** - should now show "completed" instead of "pending"

### Monitoring Logs

Look for these log entries to confirm callbacks are working:

```
# When callback is received
{"level":"info","msg":"M-Pesa callback received","remote_addr":"...","user_agent":"..."}

# When payment is processed successfully
{"level":"info","msg":"Payment completed successfully","checkout_request_id":"...","mpesa_receipt":"..."}

# When callback processing completes
{"level":"info","msg":"M-Pesa callback processed successfully","checkout_request_id":"...","status":"completed"}
```

## Deployment Instructions

### For Render Deployment

1. **Verify Environment Variables** in Render dashboard:
   ```
   MPESA_CALLBACK_URL=https://drinx-retailers-bd.onrender.com/api/v1/mpesa/callback
   ```

2. **Deploy the fixed code** - Render will auto-deploy from the connected branch

3. **Test the callback endpoint**:
   ```bash
   ./test_mpesa_callback.sh https://drinx-retailers-bd.onrender.com
   ```

4. **Monitor logs** in Render dashboard for callback activity

### For Local Development

1. **Use ngrok** to expose your local server:
   ```bash
   ngrok http 8080
   ```

2. **Update .env**:
   ```
   MPESA_CALLBACK_URL=https://your-ngrok-url.ngrok.io/api/v1/mpesa/callback
   ```

3. **Run the application**:
   ```bash
   go run cmd/main.go
   ```

4. **Test locally**:
   ```bash
   ./test_mpesa_callback.sh http://localhost:8080
   ```

## Verification Checklist

After deployment, verify:

- [ ] Callback endpoint returns 200 OK: `curl -X POST https://drinx-retailers-bd.onrender.com/api/v1/mpesa/callback -H "Content-Type: application/json" -d '{"Body":{"stkCallback":{}}}'`
- [ ] New payments transition from "pending" to "completed" within 30 seconds
- [ ] Transaction IDs are captured and displayed in frontend
- [ ] Failed payments are marked as "failed" correctly
- [ ] No stuck "pending" payments
- [ ] Logs show callback reception and processing

## Additional Notes

### Why M-Pesa Always Gets 200 OK

The callback handler **always returns HTTP 200** to M-Pesa, even when there are errors processing the callback. This is intentional to prevent M-Pesa from retrying failed callbacks, which could cause:
- Duplicate processing attempts
- Unnecessary load on the system
- Confusion in logs

Errors are logged internally but acknowledged to M-Pesa as successful deliveries.

### Callback Security

The current implementation does not authenticate callbacks. For production, consider:
1. **IP Whitelisting**: Only accept callbacks from M-Pesa's known IP addresses
2. **Request Validation**: Validate callback signatures if M-Pesa provides them
3. **Rate Limiting**: Prevent abuse of the public callback endpoint

### Troubleshooting

**Issue**: Payments still stuck in pending
- Check logs for "M-Pesa callback received" - if missing, M-Pesa isn't calling the endpoint
- Verify MPESA_CALLBACK_URL in production environment
- Confirm endpoint is publicly accessible (not behind authentication)
- Test with `curl` to ensure 200 OK response

**Issue**: Callback received but payment not updating
- Check logs for "Payment not found for CheckoutRequestID"
- Verify CheckoutRequestID is being stored during payment initiation
- Check database connectivity and transaction handling

## Files Changed

- `internals/api/routes.go` - Fixed route path
- `internals/controllers/payments.go` - Enhanced logging and error handling
- `test_mpesa_callback.sh` - New test script

## Related Issues

This fix resolves the issue where payments remain stuck in "pending" status despite successful M-Pesa completion.

## Support

For issues or questions:
- Check application logs in Render dashboard
- Review M-Pesa developer portal: https://developer.safaricom.co.ke/
- Contact M-Pesa technical support: developers@safaricom.co.ke
