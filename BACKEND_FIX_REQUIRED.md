# URGENT: Backend M-Pesa Callback Fix Required

## Issue Summary
Customers can initiate and complete M-Pesa payments successfully, but the payment status never updates from "pending" to "completed" in the database. This causes orders to remain stuck in pending status even after successful payment.

## Root Cause
The backend M-Pesa callback handler is either:
1. **Not implemented** at all
2. **Not properly configured** (incorrect URL or not publicly accessible)
3. **Implemented but not updating the database** correctly
4. **Receiving callbacks but failing silently**

## Critical Backend Changes Required

### 1. Verify M-Pesa STK Push Configuration
**Location**: Payment initiation endpoint (likely `/api/v1/payments/mpesa/initiate`)

**What to check**:
```go
// or Python/Node.js equivalent
payload := map[string]interface{}{
    "BusinessShortCode": config.MpesaShortCode,
    "Password":          generatePassword(),
    "Timestamp":         timestamp,
    "TransactionType":   "CustomerPayBillOnline",
    "Amount":            amount,
    "PartyA":            phoneNumber,
    "PartyB":            config.MpesaShortCode,
    "PhoneNumber":       phoneNumber,
    "CallBackURL":       "https://YOUR-PRODUCTION-DOMAIN.com/api/v1/payments/mpesa/callback", // ← CHECK THIS!
    "AccountReference":  orderID,
    "TransactionDesc":   "Smart Retail Payment",
}
```

**Required**:
- ✅ CallBackURL must be HTTPS (not HTTP)
- ✅ CallBackURL must be publicly accessible (not localhost, not behind VPN/firewall)
- ✅ CallBackURL must point to a real endpoint that exists
- ✅ Domain must be whitelisted with Safaricom (for production)

---

### 2. Implement M-Pesa Callback Handler

**Endpoint**: `POST /api/v1/payments/mpesa/callback`

**M-Pesa Callback Payload Structure**:
```json
{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "29115-34620561-1",
      "CheckoutRequestID": "ws_CO_191220191020363925",
      "ResultCode": 0,
      "ResultDesc": "The service request is processed successfully.",
      "CallbackMetadata": {
        "Item": [
          {
            "Name": "Amount",
            "Value": 1
          },
          {
            "Name": "MpesaReceiptNumber",
            "Value": "NLJ7RT61SV"
          },
          {
            "Name": "TransactionDate",
            "Value": 20191219102115
          },
          {
            "Name": "PhoneNumber",
            "Value": 254708374149
          }
        ]
      }
    }
  }
}
```

**Handler Implementation (Pseudo-code)**:

```python
# Python/FastAPI example
@app.post("/api/v1/payments/mpesa/callback")
async def mpesa_callback(request: Request):
    try:
        # 1. Parse the callback payload
        payload = await request.json()
        logger.info(f"M-Pesa callback received: {json.dumps(payload)}")
        
        # 2. Extract callback data
        callback_body = payload.get("Body", {}).get("stkCallback", {})
        checkout_request_id = callback_body.get("CheckoutRequestID")
        result_code = callback_body.get("ResultCode")
        result_desc = callback_body.get("ResultDesc", "")
        
        # 3. Find the order by CheckoutRequestID
        # You stored this when initiating payment
        order = db.query(Order).filter(
            Order.checkout_request_id == checkout_request_id
        ).first()
        
        if not order:
            logger.error(f"Order not found for CheckoutRequestID: {checkout_request_id}")
            # Still return 200 to M-Pesa!
            return {"ResultCode": 0, "ResultDesc": "Accepted"}
        
        # 4. Update payment status based on ResultCode
        if result_code == 0:
            # Payment successful
            callback_metadata = callback_body.get("CallbackMetadata", {}).get("Item", [])
            mpesa_receipt = next(
                (item["Value"] for item in callback_metadata if item["Name"] == "MpesaReceiptNumber"),
                None
            )
            
            order.payment_status = "completed"
            order.mpesa_transaction_id = mpesa_receipt
            order.order_status = "processing"  # Move order to processing
            logger.info(f"Order {order.id} payment completed. Receipt: {mpesa_receipt}")
            
        else:
            # Payment failed
            order.payment_status = "failed"
            logger.warning(f"Order {order.id} payment failed. ResultCode: {result_code}, Desc: {result_desc}")
        
        # 5. Commit to database
        db.commit()
        
        # 6. ALWAYS return 200 OK to M-Pesa
        return {"ResultCode": 0, "ResultDesc": "Accepted"}
        
    except Exception as e:
        logger.error(f"Error processing M-Pesa callback: {str(e)}", exc_info=True)
        # Still return 200 to prevent M-Pesa from retrying indefinitely
        return {"ResultCode": 0, "ResultDesc": "Accepted"}
```

**Critical Points**:
- ⚠️ **ALWAYS return HTTP 200** to M-Pesa, even if processing fails internally
- ⚠️ **Log everything** - callbacks, errors, database updates
- ⚠️ **Handle all exceptions** to prevent crashes
- ⚠️ **Commit database transactions** explicitly
- ⚠️ **Store CheckoutRequestID** when initiating payment to match callbacks

---

### 3. Update Order Model (If Missing Fields)

Ensure your Order/Payment model has these fields:

```python
class Order(Base):
    __tablename__ = "orders"
    
    id = Column(String, primary_key=True)
    checkout_request_id = Column(String, nullable=True, index=True)  # ← Add this!
    mpesa_transaction_id = Column(String, nullable=True)
    payment_status = Column(String, default="pending")  # pending, completed, failed
    payment_method = Column(String, default="mpesa")
    order_status = Column(String, default="processing")
    # ... other fields
```

**When initiating payment**, store the CheckoutRequestID:
```python
# After successful M-Pesa initiation
order.checkout_request_id = mpesa_response["CheckoutRequestID"]
db.commit()
```

---

### 4. Update Status Check Endpoint

**Endpoint**: `GET /api/v1/payments/{orderId}/status`

```python
@app.get("/api/v1/payments/{order_id}/status")
async def get_payment_status(order_id: str, db: Session = Depends(get_db)):
    order = db.query(Order).filter(Order.id == order_id).first()
    
    if not order:
        raise HTTPException(status_code=404, detail="Order not found")
    
    logger.info(f"Status check for order {order_id}: {order.payment_status}")
    
    return {
        "status": order.payment_status,  # pending, completed, failed
        "transactionId": order.mpesa_transaction_id or ""
    }
```

**Important**: 
- ✅ Read directly from database (no caching)
- ✅ Return latest status
- ✅ Include transaction ID if available

---

## Testing Guide

### 1. Test Callback URL Accessibility

From an **external** machine (not localhost):
```bash
curl -X POST https://your-backend-domain.com/api/v1/payments/mpesa/callback \
  -H "Content-Type: application/json" \
  -d '{
    "Body": {
      "stkCallback": {
        "MerchantRequestID": "test-123",
        "CheckoutRequestID": "test-checkout-456",
        "ResultCode": 0,
        "ResultDesc": "Test callback",
        "CallbackMetadata": {
          "Item": [
            {"Name": "Amount", "Value": 100},
            {"Name": "MpesaReceiptNumber", "Value": "TEST123XYZ"},
            {"Name": "TransactionDate", "Value": 20240123120000},
            {"Name": "PhoneNumber", "Value": 254712345678}
          ]
        }
      }
    }
  }'
```

**Expected**: HTTP 200 response

### 2. Test Database Update

1. Create a test order with a known `checkout_request_id`
2. Send test callback (above) with matching `CheckoutRequestID`
3. Query database:
```sql
SELECT id, payment_status, mpesa_transaction_id 
FROM orders 
WHERE checkout_request_id = 'test-checkout-456';
```
4. Verify `payment_status` changed to "completed"

### 3. End-to-End Test

1. Initiate real payment from frontend
2. Complete M-Pesa prompt on phone
3. Check backend logs for callback receipt:
   ```
   [INFO] M-Pesa callback received: {...}
   [INFO] Order abc123 payment completed. Receipt: NLJ7RT61SV
   ```
4. Check database for status update
5. Verify frontend detects "completed" status

---

## Debugging Checklist

### Backend Logs to Add

```python
# In payment initiation
logger.info(f"Initiating M-Pesa for order {order_id}, callback URL: {callback_url}")

# In callback handler
logger.info(f"Callback received: {json.dumps(payload)}")
logger.info(f"Extracted CheckoutRequestID: {checkout_request_id}, ResultCode: {result_code}")
logger.info(f"Found order: {order.id if order else 'NOT FOUND'}")
logger.info(f"Updated payment_status to: {order.payment_status}")

# In status check
logger.info(f"Status check for {order_id}: current status = {order.payment_status}")
```

### Common Issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Callbacks never arrive | URL not accessible | Use ngrok/tunneling for dev, proper domain for prod |
| Callbacks arrive but status not updated | Database transaction not committed | Add `db.commit()` explicitly |
| Order not found in callback | CheckoutRequestID not stored | Store it during payment initiation |
| 500 error in callback | Unhandled exception | Add try-except, log errors |
| Works in dev, not in prod | HTTP vs HTTPS | Ensure production uses HTTPS |

---

## Environment Variables to Check

```bash
# .env file
MPESA_CONSUMER_KEY=your_key
MPESA_CONSUMER_SECRET=your_secret
MPESA_SHORTCODE=174379
MPESA_PASSKEY=your_passkey
MPESA_CALLBACK_URL=https://your-production-domain.com/api/v1/payments/mpesa/callback
DATABASE_URL=postgresql://...
```

---

## Deployment Notes

### Production Requirements

1. **Domain**: Must be publicly accessible HTTPS domain
2. **Safaricom Whitelist**: Add your callback URL to Safaricom portal
3. **Server Timeout**: Ensure no timeout < 30 seconds on callback endpoint
4. **Logging**: Enable INFO level logging for payment endpoints
5. **Monitoring**: Set up alerts for:
   - Payment status stuck in "pending" > 5 minutes
   - No callbacks received in last hour (during business hours)
   - Callback errors rate > 5%

### Development/Testing

Use **ngrok** or similar for local testing:
```bash
ngrok http 8080
# Use the HTTPS URL as callback: https://abc123.ngrok.io/api/v1/payments/mpesa/callback
```

---

## Summary - What MUST Be Done

1. ✅ **Implement** `/api/v1/payments/mpesa/callback` POST endpoint
2. ✅ **Store** CheckoutRequestID when initiating payment
3. ✅ **Update** payment_status in database when callback arrives
4. ✅ **Return** HTTP 200 from callback handler always
5. ✅ **Add** comprehensive logging
6. ✅ **Ensure** callback URL is HTTPS and publicly accessible
7. ✅ **Test** with mock callback first, then real M-Pesa

---

## Timeline

- **Critical Priority**: This breaks the entire payment flow
- **Estimated Fix Time**: 2-4 hours (if callback infrastructure exists)
- **Testing Time**: 1-2 hours (including real M-Pesa test)

---

## Contact & Resources

- **M-Pesa Developer Portal**: https://developer.safaricom.co.ke/
- **M-Pesa Support**: developers@safaricom.co.ke
- **STK Push Documentation**: https://developer.safaricom.co.ke/docs#lipa-na-m-pesa-online-payment

---

## Questions to Ask Backend Team

1. Is the callback endpoint implemented? Where is the code?
2. What is the production callback URL?
3. Is the URL whitelisted with Safaricom?
4. Are callbacks being received? (Check logs)
5. Is the database being updated? (Check DB directly)
6. What errors appear in backend logs during payment?

---

**This fix MUST be done in the backend repository: https://github.com/ByteBenders-compScientists/smart-retail-backend**
