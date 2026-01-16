# Smart Retail Systems -- Backend
- Inventory management system for small retail shops

### Todos
- [x] User auth (customers and admin)
- [ ] branch inventory per branch
- [ ] sales recording (with MPESA payment integration in sandbox)
- [ ] restocking from HQ
- [ ] reporting (per brand income and totals)
- [ ] syncing for offline mode
- [ ] sms/email low-stock alerts

## Structure

```
.
├───cmd
└───internals
    ├───api
    ├───controllers
    ├───db
    ├───initialisers
    ├───middlewares
    ├───migrate
    ├───models
    └───utils
```

## Documentation

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
```

##### Get User Profile
```http
GET /api/v1/protected/profile
Authorization: Bearer YOUR_JWT_TOKEN_HERE
```

##### Admin Dashboard (Admin Only)
```http
GET /api/v1/protected/admin/dashboard
Authorization: Bearer YOUR_JWT_TOKEN_HERE
```

##### Customer Products (Customer & Admin)
```http
GET /api/v1/protected/customer/products
Authorization: Bearer YOUR_JWT_TOKEN_HERE
```

## Authors
- By [them](https://github.com/ByteBenders-compScientists)
