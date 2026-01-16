# Smart Retail Systems -- Backend
- Complete inventory management system for small retail shops with multi-branch support

## Features Implemented

### ✅ Authentication & Authorization
- JWT-based authentication for secure session management
- Role-based access control (Admin/Customer)
- User registration and login functionality

### ✅ Branch Management
- Create, view, update, and delete branches
- HQ designation for centralized management
- Branch-specific inventory tracking

### ✅ Product Management
- Complete CRUD operations for products
- Brand-based product categorization
- Stock tracking across multiple branches

### ✅ Inventory Management
- Real-time stock monitoring per branch
- Automatic stock updates during sales
- Low stock alerts and notifications
- Bulk restocking from headquarters

### ✅ Sales Processing
- Multi-item sales transactions
- Automatic stock deduction
- Payment status tracking
- Sales history and reporting

### ✅ MPESA Integration
- STK Push payment initiation (Sandbox)
- Webhook handling for payment confirmation
- Payment reference tracking

### ✅ Offline Sync
- Offline sales recording capability
- Synchronization when connectivity restored
- Duplicate transaction detection
- Conflict resolution mechanisms

### ✅ Reporting & Analytics
- Sales reports by time period (daily/weekly/monthly/yearly)
- Brand-based revenue breakdown
- Branch performance comparison
- Low stock reports
- Revenue summaries and trends

### ✅ Alert System
- Low stock notifications
- Critical stock alerts
- Branch-specific alert summaries
- Health score calculation

## API Documentation

#### Base URL
```
http://localhost:8080
```

#### Health Check
```http
GET /api/v1/health
```

#### Authentication Routes

##### Register New User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "password": "password123"
}
```

##### Login User
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john.doe@example.com",
  "password": "password123"
}
```

#### Protected Routes (Requires JWT Token)

After login, copy the token from the response and include it in the Authorization header:

```http
Authorization: Bearer YOUR_JWT_TOKEN_HERE

or

# include cookie authentications
{ credentials: 'include' }
```

### Admin Routes

#### Branch Management
```http
GET    /api/v1/admin/branches              # Get all branches
POST   /api/v1/admin/branches              # Create new branch
GET    /api/v1/admin/branches/:id          # Get specific branch
PUT    /api/v1/admin/branches/:id          # Update branch
DELETE /api/v1/admin/branches/:id          # Delete branch
GET    /api/v1/admin/branches/:id/inventory # Get branch inventory
POST   /api/v1/admin/branches/:id/stock    # Add stock to branch
PUT    /api/v1/admin/branches/:id/stock/:stockId # Update stock
GET    /api/v1/admin/branches/:id/low-stock # Get low stock alerts
```

#### Product Management
```http
GET    /api/v1/admin/products              # Get all products
POST   /api/v1/admin/products              # Create new product
GET    /api/v1/admin/products/:id          # Get specific product
PUT    /api/v1/admin/products/:id          # Update product
DELETE /api/v1/admin/products/:id          # Delete product
GET    /api/v1/admin/products/brand?brand=coke # Get products by brand
GET    /api/v1/admin/products/:id/stock    # Get product stock across branches
```

#### Sales Management
```http
POST   /api/v1/admin/branches/:id/sales   # Create sale
GET    /api/v1/admin/sales                 # Get all sales
GET    /api/v1/admin/sales/:saleId         # Get specific sale
PUT    /api/v1/admin/sales/:saleId/status  # Update sale status
GET    /api/v1/admin/sales/status?status=paid # Get sales by status
GET    /api/v1/admin/branches/:id/sales    # Get branch sales
```

#### Restocking (HQ Operations)
```http
POST   /api/v1/admin/restock               # Restock single item from HQ
POST   /api/v1/admin/restock/bulk          # Bulk restock from HQ
GET    /api/v1/admin/restock/hq-stock      # Get HQ inventory
GET    /api/v1/admin/restock/history       # Get restock history
GET    /api/v1/admin/restock/suggestions  # Get restock suggestions
```

#### Reports
```http
GET    /api/v1/admin/reports/sales?period=week # Sales report
GET    /api/v1/admin/reports/branches?period=month # Branch performance
GET    /api/v1/admin/reports/low-stock?threshold=10 # Low stock report
GET    /api/v1/admin/reports/revenue?period=month # Revenue summary
GET    /api/v1/admin/reports/trends?days=30 # Daily sales trends
```

#### Sync Management
```http
GET    /api/v1/admin/sync/pending         # Get pending sync items
PUT    /api/v1/admin/sync/:saleId/resolve  # Resolve sync conflicts
```

#### Alerts
```http
GET    /api/v1/admin/alerts/low-stock      # Get low stock alerts
GET    /api/v1/admin/alerts/critical       # Get critical stock alerts
GET    /api/v1/admin/alerts/summary        # Get alert summary
GET    /api/v1/admin/alerts/history?days=7 # Get alert history
POST   /api/v1/admin/alerts/rules          # Create alert rule
```

### Customer Routes

#### Branch & Product Viewing
```http
GET    /api/v1/customer/branches           # Get all branches
GET    /api/v1/customer/branches/:id       # Get specific branch
GET    /api/v1/customer/branches/:id/stocks # Get branch stock
GET    /api/v1/customer/branches/:id/alerts # Get branch alerts
GET    /api/v1/customer/products           # Get all products
GET    /api/v1/customer/products/:id       # Get specific product
GET    /api/v1/customer/products/brand?brand=coke # Get by brand
```

#### Sales & Payments
```http
POST   /api/v1/customer/branches/:id/sales # Create sale
GET    /api/v1/customer/sales              # Get user's sales
GET    /api/v1/customer/sales/:saleId      # Get specific sale
POST   /api/v1/customer/mpesa/initiate     # Initiate MPESA payment
POST   /api/v1/customer/sync               # Sync offline sales
GET    /api/v1/customer/sync/status?client_id=xxx # Get sync status
```

### Public Routes

#### MPESA Webhook
```http
POST   /api/v1/mpesa/webhook               # MPESA callback handler
```

## Setup Instructions

### Prerequisites
- Go 1.24.1 or higher
- PostgreSQL database
- MPESA Sandbox credentials (for payment integration)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/ByteBenders-compScientists/smart-retail-backend.git
cd smart-retail-backend
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
nano .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
go run cmd/migrate/main.go
```

5. Start the server:
```bash
go run cmd/main.go
```

6. ngrok for localhost
```bash
ngrok http 8080
```

## Environment Variables

Key environment variables required:

- `DB_URI`: Database connection
- `JWT_SECRET`: Secret key for JWT token signing
- `MPESA_CONSUMER_KEY`, `MPESA_CONSUMER_SECRET`: MPESA API credentials
- `MPESA_SHORTCODE`, `MPESA_PASSKEY`: MPESA payment configuration
- `MPESA_CALLBACK_URL`: URL for MPESA webhook callbacks

## Database Schema

The system uses the following main entities:
- **Users**: Authentication and role management
- **Branches**: Store locations with HQ designation
- **Products**: Product catalog with brand categorization
- **Stock**: Inventory levels per branch
- **Sales**: Transaction records with payment status
- **SaleItems**: Line items for each sale

## Testing

Run API tests:
```bash
bash test-api.sh
```

## Authors
- By [ByteBenders-compScientists](https://github.com/ByteBenders-compScientists)
