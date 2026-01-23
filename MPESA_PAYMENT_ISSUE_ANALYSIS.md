# M-Pesa Payment Issue Analysis & Solution

## Problem Statement
M-Pesa payments are initiated successfully and customers complete the payment on their phones, but the payment status in the system remains "pending" instead of updating to "completed".

## Root Cause Analysis

### Frontend Flow (Current Implementation)
1. **Order Creation**: User completes checkout → Creates order with `PaymentStatus='pending'`
2. **Payment Initiation**: Calls `/api/v1/payments/mpesa/initiate` → Backend initiates M-Pesa STK Push
3. **Status Polling**: Frontend polls `/api/v1/payments/{orderId}/status` every 3 seconds (max 60 attempts = 3 minutes)
4. **Expected**: Backend receives M-Pesa callback → Updates payment status → Frontend poll detects 'completed' status
5. **Actual**: Status remains 'pending' even after customer pays

### Why Frontend Polling Alone is Not the Issue
- Frontend correctly polls the backend endpoint every 3 seconds
- Frontend properly handles 'pending', 'completed', and 'failed' statuses
- The issue is that the backend is NOT updating the payment status from 'pending' to 'completed'

## Backend Issues (Probable Causes)

### 1. M-Pesa Callback URL Not Configured
**Symptom**: Backend never receives callback from M-Pesa
**Solution**: 
```
Backend needs to register a callback URL with M-Pesa Daraja API when initiating STK Push:
- CallbackURL: https://your-backend-domain.com/api/v1/payments/mpesa/callback
- Must be publicly accessible (not localhost)
- Must use HTTPS in production
```

### 2. Callback Handler Not Implemented or Not Working
**Symptom**: Callback URL is registered but handler doesn't update database
**Solution**:
Backend needs to implement POST endpoint `/api/v1/payments/mpesa/callback` that:
- Receives M-Pesa callback payload
- Extracts `ResultCode` (0 = success, non-0 = failure)
- Extracts `CheckoutRequestID` to find the order
- Updates order's `PaymentStatus` and `MpesaTransactionID`
- Logs all callbacks for debugging

### 3. Database Transaction Issues
**Symptom**: Callback handler runs but database doesn't commit changes
**Solution**:
- Ensure database transactions are committed
- Add proper error handling and rollback logic
- Log database operations

### 4. M-Pesa Callback Timing Issues
**Symptom**: Callback arrives after timeout period
**Solution**:
- M-Pesa callbacks can take 30-60 seconds
- Current frontend timeout is 3 minutes (good)
- Ensure backend has no timeout restrictions on callback endpoint

## Required Backend Changes

### 1. STK Push Initiation (Already Implemented - Verify)
```python
# When initiating payment, must include:
{
    "BusinessShortCode": "174379",
    "Password": base64_encoded_password,
    "Timestamp": timestamp,
    "TransactionType": "CustomerPayBillOnline",
    "Amount": order_amount,
    "PartyA": customer_phone,
    "PartyB": "174379",
    "PhoneNumber": customer_phone,
    "CallBackURL": "https://your-domain.com/api/v1/payments/mpesa/callback",
    "AccountReference": order_id,
    "TransactionDesc": "Smart Retail Payment"
}
```

### 2. Callback Handler Implementation (MUST BE ADDED/FIXED)
```python
POST /api/v1/payments/mpesa/callback

Expected Payload from M-Pesa:
{
    "Body": {
        "stkCallback": {
            "MerchantRequestID": "...",
            "CheckoutRequestID": "...",
            "ResultCode": 0,  # 0 = success, non-0 = failed
            "ResultDesc": "The service request is processed successfully.",
            "CallbackMetadata": {
                "Item": [
                    {"Name": "Amount", "Value": 100},
                    {"Name": "MpesaReceiptNumber", "Value": "ABC123XYZ"},
                    {"Name": "TransactionDate", "Value": 20240123143022},
                    {"Name": "PhoneNumber", "Value": 254712345678}
                ]
            }
        }
    }
}

Handler Logic:
1. Extract CheckoutRequestID or use AccountReference from request
2. Find order by CheckoutRequestID
3. If ResultCode == 0:
   - Set PaymentStatus = 'completed'
   - Set MpesaTransactionID = MpesaReceiptNumber
   - Set OrderStatus = 'processing' (if applicable)
4. Else:
   - Set PaymentStatus = 'failed'
5. Save to database with transaction
6. Return 200 OK to M-Pesa (important!)
7. Log everything for debugging
```

### 3. Status Check Endpoint (Already Implemented - Verify)
```python
GET /api/v1/payments/{orderId}/status

Should return:
{
    "status": "pending" | "completed" | "failed",
    "transactionId": "ABC123XYZ" or null
}

Must read directly from database, not cache
```

## Testing the Fix

### 1. Callback URL Accessibility
```bash
# Test from external network (not localhost)
curl -X POST https://your-backend-domain.com/api/v1/payments/mpesa/callback \
  -H "Content-Type: application/json" \
  -d '{"Body":{"stkCallback":{"ResultCode":0}}}'

# Should return 200 OK
```

### 2. Database Update Verification
```sql
-- After test payment, check database
SELECT OrderID, PaymentStatus, MpesaTransactionID, UpdatedAt 
FROM Orders 
WHERE OrderID = 'test-order-id';

-- Should show PaymentStatus = 'completed' after callback
```

### 3. End-to-End Test
1. Initiate payment from frontend
2. Complete M-Pesa payment on phone
3. Monitor backend logs for callback receipt
4. Verify database update happens
5. Confirm frontend status changes to 'completed'

## Debugging Steps

### Backend Logging (Add These)
```python
# In STK Push initiation
logger.info(f"Initiating M-Pesa payment for order {order_id}")
logger.info(f"Callback URL: {callback_url}")

# In callback handler
logger.info(f"Received M-Pesa callback: {json.dumps(request.body)}")
logger.info(f"ResultCode: {result_code}, OrderID: {order_id}")
logger.info(f"Updated PaymentStatus to {new_status}")

# In status check
logger.info(f"Status check for order {order_id}: {current_status}")
```

### Frontend Enhancement (Optional)
Add transaction ID display and better error messages (covered in frontend improvements)

## Summary Checklist

### Backend (CRITICAL - Must be fixed)
- [ ] Verify callback URL is publicly accessible HTTPS endpoint
- [ ] Implement/Fix `/api/v1/payments/mpesa/callback` handler
- [ ] Ensure handler updates `PaymentStatus` in database
- [ ] Ensure handler extracts and saves `MpesaReceiptNumber` as `MpesaTransactionID`
- [ ] Add comprehensive logging to all payment endpoints
- [ ] Test callback handler with mock M-Pesa payload
- [ ] Verify database transactions commit properly
- [ ] Ensure no request timeouts on callback endpoint

### Frontend (NICE TO HAVE - Improvements)
- [ ] Add better error messages for users
- [ ] Display transaction ID when available
- [ ] Add manual refresh button for payment status
- [ ] Show clearer pending state with expected wait time
- [ ] Add support link/contact for payment issues

## Production Deployment Notes

1. **M-Pesa Callback URL Requirements**:
   - Must be HTTPS (not HTTP)
   - Must be publicly accessible (no VPN/firewall blocking)
   - Must respond within 30 seconds
   - Must return 200 OK even if processing fails internally

2. **Environment Variables**:
   - Ensure production backend has correct M-Pesa credentials
   - Verify `MPESA_CALLBACK_URL` points to production domain

3. **Monitoring**:
   - Set up alerts for failed callbacks
   - Monitor payment status update rate
   - Track stuck payments (pending > 5 minutes)

## Contact Information

For implementation questions, contact:
- Backend Team: [Add contact]
- DevOps: [Add contact]
- M-Pesa Support: developers@safaricom.co.ke
