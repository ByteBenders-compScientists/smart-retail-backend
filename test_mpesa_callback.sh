#!/bin/bash

# Test script for M-Pesa callback endpoint
# This script tests the callback endpoint with a sample M-Pesa callback payload

set -e

# Configuration
BASE_URL="${1:-http://localhost:8080}"
CALLBACK_URL="${BASE_URL}/api/v1/mpesa/callback"

echo "==============================================="
echo "M-Pesa Callback Endpoint Test"
echo "==============================================="
echo "Target URL: $CALLBACK_URL"
echo ""

# Sample successful payment callback from M-Pesa
SUCCESSFUL_PAYLOAD='{
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
            "Value": 1.00
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
}'

# Sample failed payment callback from M-Pesa
FAILED_PAYLOAD='{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "29115-34620561-2",
      "CheckoutRequestID": "ws_CO_191220191020363926",
      "ResultCode": 1032,
      "ResultDesc": "Request cancelled by user"
    }
  }
}'

echo "Test 1: Successful Payment Callback"
echo "-----------------------------------"
HTTP_CODE=$(curl -s -o /tmp/callback_response.txt -w "%{http_code}" \
  -X POST "$CALLBACK_URL" \
  -H "Content-Type: application/json" \
  -d "$SUCCESSFUL_PAYLOAD")

echo "HTTP Status Code: $HTTP_CODE"
echo "Response:"
cat /tmp/callback_response.txt
echo ""
echo ""

if [ "$HTTP_CODE" = "200" ]; then
  echo "✅ Test 1 PASSED - Successful callback accepted"
else
  echo "❌ Test 1 FAILED - Expected HTTP 200, got $HTTP_CODE"
  exit 1
fi

echo ""
echo "Test 2: Failed Payment Callback"
echo "--------------------------------"
HTTP_CODE=$(curl -s -o /tmp/callback_response.txt -w "%{http_code}" \
  -X POST "$CALLBACK_URL" \
  -H "Content-Type: application/json" \
  -d "$FAILED_PAYLOAD")

echo "HTTP Status Code: $HTTP_CODE"
echo "Response:"
cat /tmp/callback_response.txt
echo ""
echo ""

if [ "$HTTP_CODE" = "200" ]; then
  echo "✅ Test 2 PASSED - Failed callback accepted"
else
  echo "❌ Test 2 FAILED - Expected HTTP 200, got $HTTP_CODE"
  exit 1
fi

echo ""
echo "Test 3: Invalid Payload (should still return 200)"
echo "------------------------------------------------"
HTTP_CODE=$(curl -s -o /tmp/callback_response.txt -w "%{http_code}" \
  -X POST "$CALLBACK_URL" \
  -H "Content-Type: application/json" \
  -d '{"invalid": "payload"}')

echo "HTTP Status Code: $HTTP_CODE"
echo "Response:"
cat /tmp/callback_response.txt
echo ""
echo ""

if [ "$HTTP_CODE" = "200" ]; then
  echo "✅ Test 3 PASSED - Invalid callback still returns 200"
else
  echo "❌ Test 3 FAILED - Expected HTTP 200, got $HTTP_CODE"
  exit 1
fi

echo ""
echo "==============================================="
echo "All Tests Passed! ✅"
echo "==============================================="
echo ""
echo "Note: These tests verify the endpoint is accessible and"
echo "returns proper responses. For full testing, you need:"
echo "  1. A database with actual payment records"
echo "  2. Valid CheckoutRequestID values in the test payloads"
echo ""
echo "To test with production/render:"
echo "  ./test_mpesa_callback.sh https://drinx-retailers-bd.onrender.com"
echo ""
