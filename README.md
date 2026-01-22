# Smart Retail Backend - Drinx Retailers System

Complete distributed supermarket management system backend with multi-branch support, real-time inventory tracking, and M-Pesa payment integration.

## 🚀 Latest Updates

### ✅ **Frontend Integration Complete**
- **Product Model**: Updated to match frontend with all fields (name, brand, description, price, originalPrice, image, rating, reviews, category, volume, unit, tags)
- **12 Products**: Complete product catalog with Coke, Fanta, Sprite variants (single bottles, crates, different volumes)
- **Tags System**: JSON-based product tags for filtering and categorization
- **Real Images**: Updated with proper image URLs for all products

### ✅ **Order System Refactored**
- **Order/OrderItem Models**: Replaced old Sale/SaleItem system
- **Payment Integration**: Full M-Pesa STK Push with webhook handling
- **Branch Inventory**: Real-time stock tracking per branch
- **Audit Trail**: RestockLog for complete inventory history

### ✅ **M-Pesa Payment System**
- **Dual Integration**: Both new Order system and legacy Sale system supported
- **Real STK Push**: Actual Safaricom sandbox integration
- **Webhook Handling**: Proper callback processing with transaction updates
- **Payment Status**: Complete payment lifecycle tracking

## 🏗️ Architecture Overview

### **Database Models**
- **User**: Authentication with phone validation, role-based access (admin/customer)
- **Branch**: 5 predefined locations (Nairobi HQ, Kisumu, Mombasa, Nakuru, Eldoret)
- **Product**: 12 products with detailed fields and tags
- **BranchInventory**: Stock tracking per branch with low-stock alerts
- **Order/OrderItem**: Order management with payment status
- **Payment**: M-Pesa transaction tracking
- **RestockLog**: Audit trail for all inventory movements

### **Key Features**
- 🔐 **JWT Authentication** with role-based access control
- 📦 **Multi-Branch Inventory** with real-time tracking
- 💳 **M-Pesa Integration** with STK Push and callbacks
- 📊 **Comprehensive Reporting** by brand, branch, and time
- 🚨 **Low Stock Alerts** with threshold-based notifications
- 📱 **Phone Validation** for user registration
- 🏷️ **Product Tags** for enhanced filtering
- 📋 **Order Management** with complete payment lifecycle

## 📡 API Documentation

### **Base URL**
```
http://localhost:8080
```

### **Authentication Routes**

#### **Register User**
```http
POST /api/v1/auth/register
```
**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+254712345678",
  "password": "securepassword123",
  "confirmPassword": "securepassword123"
}
```
**Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "uuid-here",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+254712345678",
    "role": "customer"
  },
  "token": "jwt-token-here"
}
```

#### **Login**
```http
POST /api/v1/auth/login
```
**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
```
**Response (200 OK):**
```json
{
  "message": "Login successful",
  "user": {
    "id": "uuid-here",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+254712345678",
    "role": "customer"
  },
  "token": "jwt-token-here"
}
```

#### **Logout**
```http
POST /api/v1/auth/logout
```
**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

#### **Get Current User**
```http
GET /api/v1/auth/me
```
**Headers:**
```
Authorization: Bearer jwt-token-here
```
**Response (200 OK):**
```json
{
  "id": "uuid-here",
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+254712345678",
  "role": "customer",
  "createdAt": "2026-01-22T10:30:00Z"
}
```

### **Product Routes**

#### **List All Products**
```http
GET /api/v1/products
```
**Response (200 OK):**
```json
[
  {
    "id": "uuid-here",
    "name": "Coca-Cola Original 500ml",
    "brand": "Coke",
    "description": "Classic Coca-Cola taste. Available in single bottles or crates.",
    "price": 60.00,
    "originalPrice": 65.00,
    "image": "https://i.postimg.cc/y6SN9pt5/coke.png",
    "rating": 4.8,
    "reviews": 1250,
    "category": "Soft Drinks",
    "volume": "500ml",
    "unit": "single",
    "tags": "[\"popular\", \"classic\", \"carbonated\", \"original\"]"
  }
]
```

#### **Get Product Details**
```http
GET /api/v1/products/:id
```
**Response (200 OK):**
```json
{
  "id": "uuid-here",
  "name": "Coca-Cola Original 500ml",
  "brand": "Coke",
  "description": "Classic Coca-Cola taste. Available in single bottles or crates.",
  "price": 60.00,
  "originalPrice": 65.00,
  "image": "https://i.postimg.cc/y6SN9pt5/coke.png",
  "rating": 4.8,
  "reviews": 1250,
  "category": "Soft Drinks",
  "volume": "500ml",
  "unit": "single",
  "tags": "[\"popular\", \"classic\", \"carbonated\", \"original\"]"
}
```

#### **Get Branch Inventory with Stock**
```http
GET /api/v1/products/branch/:branchId
```
**Response (200 OK):**
```json
[
  {
    "id": "uuid-here",
    "name": "Coca-Cola Original 500ml",
    "brand": "Coke",
    "description": "Classic Coca-Cola taste.",
    "price": 60.00,
    "originalPrice": 65.00,
    "image": "https://i.postimg.cc/y6SN9pt5/coke.png",
    "rating": 4.8,
    "reviews": 1250,
    "stock": 100,
    "category": "Soft Drinks",
    "volume": "500ml",
    "unit": "single",
    "tags": ["popular", "classic", "carbonated", "original"],
    "available": true
  }
]
```

#### **Create Product (Admin)**
```http
POST /api/v1/admin/products
```
**Request Body:**
```json
{
  "name": "New Product",
  "brand": "Coke",
  "description": "Product description",
  "price": 60.00,
  "originalPrice": 65.00,
  "image": "https://example.com/image.png",
  "rating": 4.5,
  "reviews": 100,
  "category": "Soft Drinks",
  "volume": "500ml",
  "unit": "single",
  "tags": ["new", "popular"]
}
```
**Response (201 Created):**
```json
{
  "id": "uuid-here",
  "name": "New Product",
  "brand": "Coke",
  "price": 60.00,
  "originalPrice": 65.00,
  "category": "Soft Drinks"
}
```

#### **Update Product (Admin)**
```http
PUT /api/v1/admin/products/:id
```
**Request Body:**
```json
{
  "price": 55.00,
  "originalPrice": 65.00,
  "description": "Updated description"
}
```
**Response (200 OK):**
```json
{
  "id": "uuid-here",
  "name": "Product Name",
  "price": 55.00,
  "originalPrice": 65.00
}
```

#### **Delete Product (Admin)**
```http
DELETE /api/v1/admin/products/:id
```
**Response (200 OK):**
```json
{
  "message": "Product deleted successfully"
}
```

### **Branch Routes**

#### **List All Branches**
```http
GET /api/v1/branches
```
**Response (200 OK):**
```json
[
  {

#### **Create New Order**
```http
POST /api/v1/orders
```
**Headers:**
```
Authorization: Bearer jwt-token-here
```
**Request Body:**
```json
{
  "branchId": "branch-nairobi",
  "items": [
    {
      "productId": "uuid-product-1",
      "productBrand": "Coke",
      "quantity": 2,
      "price": 60.00,
      "subtotal": 120.00
    },
    {
      "productId": "uuid-product-2",
      "productBrand": "Fanta",
      "quantity": 1,
      "price": 60.00,
      "subtotal": 60.00
    }
  ],
  "totalAmount": 180.00,
  "phone": "+254712345678"
}
```
**Response (201 Created):**
```json
{
  "order": {
    "id": "uuid-order-id",
    "userId": "uuid-user-id",
    "branchId": "branch-nairobi",
    "totalAmount": 180.00,
    "paymentStatus": "pending",
    "paymentMethod": "mpesa",
    "orderStatus": "processing",
    "createdAt": "2026-01-22T10:00:00Z"
  },
  "paymentUrl": "/api/payments/mpesa/initiate"
}
```

#### **Get User Orders**
```http
GET /api/v1/orders
```
**Headers:**
```
Authorization: Bearer jwt-token-here
```
**Response (200 OK):**
```json
[
  {
    "id": "uuid-order-id",
    "userId": "uuid-user-id",
    "branchId": "branch-nairobi",
    "branch": {
      "id": "branch-nairobi",
      "name": "Nairobi"
    },
    "totalAmount": 180.00,
    "paymentStatus": "completed",
    "paymentMethod": "mpesa",
    "orderStatus": "completed",
    "orderItems": [
      {
        "id": "uuid-item-id",
        "productId": "uuid-product-1",
        "product": {
          "id": "uuid-product-1",
          "name": "Coca-Cola Original 500ml",
          "brand": "Coke",
          "price": 60.00
        },
        "productBrand": "Coke",
        "quantity": 2,
        "price": 60.00,
        "subtotal": 120.00
      }
    ],
    "createdAt": "2026-01-22T10:00:00Z",
    "completedAt": "2026-01-22T10:05:00Z"
  }
]
```

#### **Get Order Details**
```http
GET /api/v1/orders/:id
```

#### **Initiate M-Pesa Payment**
```http
POST /api/v1/payments/mpesa/initiate
```
**Headers:**
```

#### **Restock Branch Inventory**
```http
POST /api/v1/admin/restock
```
**Headers:**
```
Authorization: Bearer jwt-token-here (admin role required)
```
**Request Body:**
```json
{
  "branchId": "branch-nairobi",
  "productId": "uuid-product-id",
  "quantity": 50
}
```
**Response (200 OK):**
```json
{
  "message": "Branch restocked successfully",
  "updatedInventory": {
    "branchId": "branch-nairobi",
    "productId": "uuid-product-id",
    "previousQty": 100,
    "addedQty": 50,
    "newQty": 150
  }
}
```

#### **Get All Inventory with Alerts**
```http
GET /api/v1/admin/inventory?branchId=branch-nairobi
```
**Headers:**
```
Authorization: Bearer jwt-token-here (admin role required)
```
**Query Parameters:**
- `branchId` (optional): Filter by specific branch

**Response (200 OK):**
```json
{
  "inventory": [
    {
      "id": "uuid-inventory-id",
      "branchId": "branch-nairobi",
      "branch": {
        "id": "branch-nairobi",
        "name": "Nairobi"
      },
      "productId": "uuid-product-id",
      "product": {
        "id": "uuid-product-id",
        "name": "Coca-Cola Original 500ml",
        "brand": "Coke"
      },
      "quantity": 15,
      "lastRestocked": "2026-01-22T10:00:00Z"
    }
  ],
  "lowStockAlerts": [
    {
      "branchId": "branch-nairobi",
      "branchName": "Nairobi",
      "productId": "uuid-product-id",
      "productName": "Coca-Cola Original 500ml",
      "currentStock": 15,
      "threshold": 20
    }
  ],
  "lowStockThreshold": 20
}
```

#### **Get Restock History**
```http
GET /api/v1/admin/restock-logs?branchId=branch-nairobi&startDate=2026-01-01&endDate=2026-01-31
```
**Headers:**
```
Authorization: Bearer jwt-token-here (admin role required)
```
**Query Parameters:**
- `branchId` (optional): Filter by specific branch
- `startDate` (optional): Start date filter
- `endDate` (optional): End date filter

**Response (200 OK):**
```json
[
  {
    "id": "uuid-log-id",
    "branchId": "branch-nairobi",
    "branch": {
      "name": "Nairobi"
    },
    "productId": "uuid-product-id",
    "product": {
      "name": "Coca-Cola Original 500ml"
    },
    "quantityAdded": 50,
    "previousQuantity": 100,
    "newQuantity": 150,
    "restockedBy": "uuid-admin-id",
    "restockedByUser": {
      "name": "Admin User",
      "email": "admin@drinx.com"
    },
    "createdAt": "2026-01-22T10:00:00Z"
  }
]
```

#### **Sales Reports**
```http
GET /api/v1/admin/reports/sales?startDate=2026-01-01&endDate=2026-01-31&branchId=branch-nairobi
```
**Headers:**
```
Authorization: Bearer jwt-token-here (admin role required)
```
**Query Parameters:**
- `startDate` (optional): Start date filter
- `endDate` (optional): End date filter
- `branchId` (optional): Filter by specific branch
- `productId` (optional): Filter by specific product

**Response (200 OK):**
```json
{
  "salesByBrand": {
    "Coke": {
      "units": 500,
      "revenue": 30000.00
    },
    "Fanta": {
      Error Responses**

All endpoints may return the following error responses:

**400 Bad Request:**
```json
{
  "error": "Invalid request format"
}
```

**401 Unauthorized:**
```json
{
  "error": "User not authenticated"
}
```

**403 Forbidden:**
```json
{
  "error": "Admin access required"
}
```

**404 Not Found:**
```json
{
  "error": "Resource not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Internal server error message"
}
      "units": 200,
      "revenue": 12000.00
    }
  },
  "salesByBranch": {
    "Nairobi": 35000.00,
    "Kisumu": 15000.00,
    "Mombasa": 10000.00
  },
  "grandTotal": 60000.00,
  "filters": {
    "startDate": "2026-01-01",
    "endDate": "2026-01-31",
    "branchId": "",
    "productId": ""
  }
}
```

#### **Branch-Specific Reports**
```http
GET /api/v1/admin/reports/branch/:branchId?startDate=2026-01-01&endDate=2026-01-31
```
**Headers:**
```
Authorization: Bearer jwt-token-here (admin role required)
```
**Query Parameters:**
- `startDate` (optional): Start date filter
- `endDate` (optional): End date filter

**Response (200 OK):**
```json
{
  "branch": {
    "id": "branch-nairobi",
    "name": "Nairobi",
    "isHeadquarter": true,
    "address": "Nairobi CBD, Kenya"
  },
  "branchSales": {
    "totalRevenue": 35000.00,
    "totalOrders": 150,
    "topProducts": {
      "Coke": 250,
      "Fanta": 120,
      "Sprite": 80
    }
  }
}
  "phone": "254712345678",
  "amount": 180.00
}
```
**Response (200 OK):**
```json
{
  "success": true,
  "message": "Payment initiated successfully",
  "transactionId": "MPESA_ws_CO_22012026103045123456",
  "checkoutRequestId": "ws_CO_22012026103045123456",
  "merchantRequestId": "29115-34620561-1"
}
```

#### **Check Payment Status**
```http
GET /api/v1/payments/:orderId/status
```
**Headers:**
```
Authorization: Bearer jwt-token-here
```
**Response (200 OK):**
```json
{
  "status": "completed",
  "transactionId": "QAX12345"
}
```
**Or if pending:**
```json
{
  "status": "pending"
}
```

#### **M-Pesa Webhook (No Auth Required)**
```http
POST /api/v1/payments/mpesa/callback
```
**Request Body (from Safaricom):**
```json
{
  "Body": {
    "stkCallback": {
      "MerchantRequestID": "29115-34620561-1",
      "CheckoutRequestID": "ws_CO_22012026103045123456",
      "ResultCode": 0,
      "ResultDesc": "The service request is processed successfully.",
      "CallbackMetadata": {
        "Item": [
          {
            "Name": "Amount",
            "Value": 180
          },
          {
            "Name": "MpesaReceiptNumber",
            "Value": "QAX12345"
          },
          {
            "Name": "TransactionDate",
            "Value": 20260122103045
          },
          {
            "Name": "PhoneNumber",
            "Value": 254712345678
          }
        ]
      }
    }
  }
}
```
**Response (200 OK):**
```json
{
  "success": true
}
**Response (200 OK):**
```json
{
  "id": "uuid-order-id",
  "userId": "uuid-user-id",
  "branchId": "branch-nairobi",
  "branch": {
    "id": "branch-nairobi",
    "name": "Nairobi",
    "address": "Nairobi CBD, Kenya"
  },
  "totalAmount": 180.00,
  "paymentStatus": "completed",
  "paymentMethod": "mpesa",
  "mpesaTransactionId": "QAX12345",
  "orderStatus": "completed",
  "orderItems": [
    {
      "id": "uuid-item-id",
      "productId": "uuid-product-1",
      "product": {
        "name": "Coca-Cola Original 500ml",
        "brand": "Coke"
      },
      "quantity": 2,
      "price": 60.00,
      "subtotal": 120.00
    }
  ],
  "createdAt": "2026-01-22T10:00:00Z",
  "completedAt": "2026-01-22T10:05:00Z"
}
    "phone": "+254 700 000 001",
    "status": "active",
    "createdAt": "2026-01-22T10:00:00Z",
    "updatedAt": "2026-01-22T10:00:00Z"
  }
]
```

#### **Get Branch Details**
```http
GET /api/v1/branches/:id
```
**Response (200 OK):**
```json
{
  "id": "branch-nairobi",
  "name": "Nairobi",
  "isHeadquarter": true,
  "address": "Nairobi CBD, Kenya",
  "phone": "+254 700 000 001",
  "status": "active",
  "createdAt": "2026-01-22T10:00:00Z",
  "updatedAt": "2026-01-22T10:00:00Z"
}
```

#### **Create Branch (Admin)**
```http
POST /api/v1/admin/branches
```
**Request Body:**
```json
{
  "id": "branch-eldoret",
  "name": "Eldoret",
  "isHeadquarter": false,
  "address": "Eldoret Town, Kenya",
  "phone": "+254 700 000 005",
  "status": "active"
}
```
**Response (201 Created):**
```json
{
  "id": "branch-eldoret",
  "name": "Eldoret",
  "isHeadquarter": false,
  "address": "Eldoret Town, Kenya",
  "phone": "+254 700 000 005",
  "status": "active"
}
```

#### **Update Branch (Admin)**
```http
PUT /api/v1/admin/branches/:id
```
**Request Body:**
```json
{
  "address": "New Address, Kenya",
  "phone": "+254 700 000 999",
  "status": "active"
}
```
**Response (200 OK):**
```json
{
  "id": "branch-eldoret",
  "name": "Eldoret",
  "address": "New Address, Kenya",
  "phone": "+254 700 000 999",
  "status": "active"
}
```

#### **Delete Branch (Admin)**
```http
DELETE /api/v1/admin/branches/:id
```
**Response (200 OK):**
```json
{
  "message": "Branch deleted successfully"
}
```

### **Order Routes**
```http
POST /api/v1/orders          # Create new order
GET  /api/v1/orders          # Get user orders
GET  /api/v1/orders/:id      # Get order details
```

### **Payment Routes**
```http
POST /api/v1/payments/mpesa/initiate     # Initiate M-Pesa payment
GET  /api/v1/payments/:orderId/status   # Check payment status
POST /api/v1/payments/mpesa/callback     # M-Pesa webhook (no auth)
```

### **Admin Routes**
```http
POST /api/v1/admin/restock           # Restock branch inventory
GET  /api/v1/admin/inventory          # Get all inventory with alerts
GET  /api/v1/admin/restock-logs       # Get restock history
GET  /api/v1/admin/reports/sales      # Sales reports
GET  /api/v1/admin/reports/branch/:id # Branch-specific reports
```

### **Legacy M-Pesa Routes** (Backward Compatible)
```http
POST /api/v1/mpesa/initiate    # Initiate payment (old system)
POST /api/v1/mpesa/webhook     # M-Pesa callback (old system)
```

## 🛠️ Setup Instructions

### **Prerequisites**
- Go 1.24.1 or higher
- PostgreSQL database
- M-Pesa Sandbox credentials (for payment integration)

### **Installation**

1. **Clone the repository:**
```bash
git clone https://github.com/ByteBenders-compScientists/smart-retail-backend.git
cd smart-retail-backend
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Set up environment variables:**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Run database migrations and seed data:**
```bash
go run ./internals/migrate
```

5. **Start the server:**
```bash
go run cmd/main.go
```

6. **Setup ngrok for local development (optional):**
```bash
ngrok http 8080
# Update MPESA_CALLBACK_URL with ngrok URL
```

### **Environment Variables**

Required environment variables:

```env
# Database
DB_URI=postgresql://user:password@localhost:5432/drinx-retail

# JWT Authentication
JWT_SECRET=your_jwt_secret_key
JWT_COOKIE_NAME=auth_token
COOKIE_SECURE=false

# M-Pesa Integration
MPESA_CONSUMER_KEY=your_mpesa_consumer_key
MPESA_CONSUMER_SECRET=your_mpesa_consumer_secret
MPESA_SHORTCODE=174379
MPESA_PASSKEY=your_mpesa_passkey
MPESA_CALLBACK_URL=https://yourdomain.com/api/v1/payments/mpesa/callback
```

## 🗄️ Database Schema

### **Core Models**
- **Users**: Authentication with phone validation and role management
- **Branches**: 5 predefined locations with HQ designation
- **Products**: 12 products with comprehensive fields and tags
- **BranchInventory**: Stock tracking per branch with low-stock alerts
- **Orders**: Order management with payment status tracking
- **OrderItems**: Detailed line items for each order
- **Payments**: M-Pesa transaction tracking and status
- **RestockLogs**: Complete audit trail for inventory movements

### **Seeded Data**
The migration automatically creates:
- **5 Branches**: Nairobi HQ + 4 regional branches
- **12 Products**: Complete Coke, Fanta, Sprite catalog
- **Initial Inventory**: Nairobi (100 units), Others (50 units)
- **Admin User**: admin@drinx.com / password: password

## 🧪 Testing

### **API Testing**
```bash
# Run comprehensive API tests
bash test-api.sh
```

### **M-Pesa Testing**
- Uses Safaricom sandbox environment
- Test amounts automatically set to KSh 5
- No real money transactions

## 📊 Product Catalog

### **Available Products**
1. **Coca-Cola Original 500ml** - KSh 60.00
2. **Coca-Cola Original 500ml Crate** - KSh 1,400.00 (24 bottles)
3. **Coca-Cola Original 1 Litre** - KSh 110.00
4. **Fanta Orange 500ml** - KSh 60.00
5. **Fanta Orange 500ml Crate** - KSh 1,400.00 (24 bottles)
6. **Fanta Orange 2 Litre** - KSh 180.00
7. **Sprite Lemon-Lime 500ml** - KSh 60.00
8. **Sprite Lemon-Lime 500ml Crate** - KSh 1,400.00 (24 bottles)
9. **Coca-Cola Zero Sugar 500ml** - KSh 65.00
10. **Sprite Zero Sugar 1 Litre** - KSh 115.00
11. **Fanta Pineapple 500ml** - KSh 60.00
12. **Coca-Cola Vanilla 500ml** - KSh 70.00 (Limited Edition)

### **Product Categories**
- **Soft Drinks** (8 products)
- **Diet Drinks** (2 products)
- **Special Editions** (1 product)

### **Branch Locations**
- **Nairobi HQ** - Main warehouse and flagship store
- **Kisumu Branch** - Western region operations
- **Mombasa Branch** - Coastal region operations
- **Nakuru Branch** - Central region operations
- **Eldoret Branch** - North Rift region operations

## 🔄 Migration Guide

### **From Legacy System**
The backend supports both legacy Sale system and new Order system:
- **Legacy endpoints** remain functional for backward compatibility
- **New endpoints** provide enhanced features and better data structure
- **Gradual migration** possible without downtime

### **Key Changes**
- **Sale → Order**: Enhanced order management with payment lifecycle
- **Stock → BranchInventory**: Better inventory tracking with alerts
- **Product Fields**: Complete frontend compatibility with tags
- **Payment System**: Real M-Pesa integration with webhook handling

## 🚀 Deployment

### **Production Setup**
1. Set up PostgreSQL database
2. Configure environment variables
3. Run migrations: `go run internals/migrate/migrate.go`
4. Build binary: `go build -o smart-retail cmd/main.go`
5. Deploy with process manager (systemd/supervisor)

### **Docker Support**
```bash
# Build image
docker build -t smart-retail-backend .

# Run container
docker run -p 8080:8080 --env-file .env smart-retail-backend
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👥 Authors

- **[ByteBenders-compScientists](https://github.com/ByteBenders-compScientists)**

---

## 📞 Support

For support and questions:
- Create an issue in the repository
- Contact the development team
- Check the API documentation above
