# emotionalBeach API Documentation - English Version

## 📌 Documentation Overview

- **Project Name**: emotionalBeach
- **API Version**: v1.0
- **Base URL**: `http://localhost:8080`
- **Authentication**: JWT Token (Bearer Token)
- **Default Port**: 8080

---

## 🔑 Authentication

### JWT Token Acquisition

1. Users obtain a Token after registration or login
2. Add to request header for all authenticated endpoints: `Authorization: <token>`

### Token Validity

- **Validity Period**: 7 days
- **After Expiration**: Re-login required to obtain new Token

---

## 📍 API Endpoints

### I. User Authentication Endpoints

#### 1.1 User Registration

**Endpoint**: `POST /register`

**Authentication**: ❌ Not Required

**Request Type**: `multipart/form-data`

**Request Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | ✅ | Username |
| password | string | ✅ | Password |
| repeat_password | string | ✅ | Confirm password (must match password) |
| phone | string | ✅ | Phone number (11 digits) |
| email | string | ❌ | Email address |

**Request Example**:

```bash
curl -X POST http://localhost:8080/register \
  -F "name=John Doe" \
  -F "password=123456" \
  -F "repeat_password=123456" \
  -F "phone=13800138000" \
  -F "email=user@example.com"
```

**Response Example**:

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "name": "John Doe",
    "phone": "13800138000",
    "email": "user@example.com",
    "avatar": "",
    "gender": "male",
    "role": "user"
  },
  "message": "User registered successfully!"
}
```

**Error Handling**:

| Error Code | Description |
|-----------|-------------|
| 401 | Username, password, or confirm password cannot be empty |
| 401 | Passwords do not match |
| 401 | Phone number cannot be empty |
| 401 | Phone number must be 11 digits |
| 403 | Invalid phone number |
| 401 | User already registered |

---

#### 1.2 User Login

**Endpoint**: `POST /login`

**Authentication**: ❌ Not Required

**Request Type**: `application/json`

**Request Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| username | string | ✅ | Username |
| password | string | ✅ | Password |

**Request Example**:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "John Doe",
    "password": "123456"
  }'
```

**Response Example**:

```json
{
  "code": 200,
  "message": "Login successful",
  "user_id": 1,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Handling**:

| Error Code | Description |
|-----------|-------------|
| 401 | Invalid request format |
| 403 | Login failed |
| 404 | Username does not exist |
| 401 | Wrong password |

---

#### 1.3 GitHub OAuth Login

**Endpoint**: `GET /login/github`

**Authentication**: ❌ Not Required

**Function**: Redirect to GitHub authorization page

**Request Example**:

```bash
curl -X GET http://localhost:8080/login/github
```

**Description**:
- User is redirected to GitHub OAuth authorization page
- After authorization, callback to `/callback` endpoint
- Finally redirects to Swagger documentation page

---

#### 1.4 GitHub Callback Endpoint

**Endpoint**: `GET /callback`

**Authentication**: ❌ Not Required

**Query Parameters**:

| Parameter | Description |
|-----------|-------------|
| code | GitHub authorization code |

**Function**: Handle GitHub OAuth callback, fetch user information, and generate Token

**Description**: 
- This endpoint is called automatically by GitHub
- Redirects to Swagger documentation on success

---

### II. User Management Endpoints

#### 2.1 Get All Users

**Endpoint**: `GET /v1/user/list`

**Authentication**: ✅ JWT Token Required

**Request Header**:

```
Authorization: <token>
```

**Request Example**:

```bash
curl -X GET http://localhost:8080/v1/user/list \
  -H "Authorization: your_token_here"
```

**Response Example**:

```json
{
  "code": 200,
  "data": [
    {
      "name": "John Doe",
      "avatar": "",
      "gender": "male",
      "phone": "13800138000",
      "email": "user@example.com",
      "identity": ""
    },
    {
      "name": "Jane Smith",
      "avatar": "",
      "gender": "female",
      "phone": "13900139000",
      "email": "user2@example.com",
      "identity": ""
    }
  ],
  "message": "success"
}
```

---

#### 2.2 Conditional User Query

**Endpoint**: `GET /v1/user/condition`

**Authentication**: ✅ JWT Token Required

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | ❌ | User ID |
| email | string | ❌ | Email address |
| phone | string | ❌ | Phone number |

**Note**: At least one query parameter must be provided

**Request Example**:

```bash
# Query by ID
curl -X GET "http://localhost:8080/v1/user/condition?id=1" \
  -H "Authorization: your_token_here"

# Query by email
curl -X GET "http://localhost:8080/v1/user/condition?email=user@example.com" \
  -H "Authorization: your_token_here"

# Query by phone
curl -X GET "http://localhost:8080/v1/user/condition?phone=13800138000" \
  -H "Authorization: your_token_here"
```

**Response Example**:

```json
{
  "code": 200,
  "data": {
    "name": "John Doe",
    "avatar": "",
    "gender": "male",
    "phone": "13800138000",
    "email": "user@example.com",
    "identity": ""
  },
  "message": "success"
}
```

**Error Handling**:

| Error Code | Description |
|-----------|-------------|
| 400 | At least one parameter required (id, email, or phone) |
| 500 | Query failed |

---

#### 2.3 Update User Information

**Endpoint**: `POST /v1/user/update`

**Authentication**: ✅ JWT Token Required

**Request Type**: `multipart/form-data`

**Request Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | string | ✅ | User ID |
| name | string | ❌ | Username |
| password | string | ❌ | New password |
| phone | string | ❌ | Phone number |
| email | string | ❌ | Email address |
| avatar | string | ❌ | Avatar URL |
| gender | string | ❌ | Gender (male/female) |

**Request Example**:

```bash
curl -X POST http://localhost:8080/v1/user/update \
  -H "Authorization: your_token_here" \
  -F "id=1" \
  -F "name=John New" \
  -F "avatar=https://example.com/avatar.jpg" \
  -F "gender=male"
```

**Response Example**:

```json
{
  "code": 200,
  "data": {
    "uid": 1
  },
  "message": "success"
}
```

**Error Handling**:

| Error Code | Description |
|-----------|-------------|
| 500 | Update failed |

---

#### 2.4 Delete User

**Endpoint**: `DELETE /v1/user/delete`

**Authentication**: ✅ JWT Token Required

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| id | uint | ✅ | User ID |

**Request Example**:

```bash
curl -X DELETE "http://localhost:8080/v1/user/delete?id=1" \
  -H "Authorization: your_token_here"
```

**Response Example**:

```json
{
  "code": 200,
  "message": "success"
}
```

---

### III. Friend Relationship Endpoints

#### 3.1 Get Friend List

**Endpoint**: `POST /v1/relation/list`

**Authentication**: ✅ JWT Token Required

**Request Type**: `multipart/form-data`

**Request Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| userId | uint | ✅ | User ID |

**Request Example**:

```bash
curl -X POST http://localhost:8080/v1/relation/list \
  -H "Authorization: your_token_here" \
  -F "userId=1"
```

**Response Example**:

```json
{
  "code": 200,
  "data": {
    "count": 2,
    "users": [
      {
        "name": "Jane Smith",
        "avatar": "",
        "gender": "female",
        "phone": "13900139000",
        "email": "user2@example.com",
        "identity": ""
      },
      {
        "name": "Bob Johnson",
        "avatar": "",
        "gender": "male",
        "phone": "14000140000",
        "email": "user3@example.com",
        "identity": ""
      }
    ]
  },
  "message": "success"
}
```

**Error Handling**:

| Error Code | Description |
|-----------|-------------|
| 404 | No friends found |

---

#### 3.2 Add Friend

**Endpoint**: `POST /v1/relation/add`

**Authentication**: ✅ JWT Token Required

**Request Type**: `multipart/form-data`

**Request Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| userId | uint | ✅ | Current user ID |
| targetName | string/uint | ✅ | Target user ID or username |

**Request Example**:

```bash
# Add by user ID
curl -X POST http://localhost:8080/v1/relation/add \
  -H "Authorization: your_token_here" \
  -F "userId=1" \
  -F "targetName=2"

# Add by username
curl -X POST http://localhost:8080/v1/relation/add \
  -H "Authorization: your_token_here" \
  -F "userId=1" \
  -F "targetName=Jane Smith"
```

**Response Example**:

```json
{
  "code": 200,
  "data": {
    "msg": "Friend added successfully"
  },
  "message": "success"
}
```

---

### IV. Webhook Email Endpoint

#### 4.1 Send Email Notification

**Endpoint**: `POST /v1/api/webhook`

**Authentication**: ✅ JWT Token Required

**Request Type**: `application/json`

**Request Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| title | string | ✅ | Email subject |
| content | string | ❌ | Email content (supports HTML) |
| receivers | string[] | ✅ | List of recipient emails |

**Request Example**:

```bash
curl -X POST http://localhost:8080/v1/api/webhook \
  -H "Authorization: your_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Welcome to emotionalBeach",
    "content": "<h1>Welcome!</h1><p>This is an HTML email</p>",
    "receivers": [
      "user@example.com",
      "another@example.com"
    ]
  }'
```

**Response Example**:

```json
{
  "code": 200,
  "data": {
    "message": "Email sent successfully"
  },
  "message": "success"
}
```

**Email Features**:
- ✅ Multiple recipients support
- ✅ HTML content support
- ✅ Automatic HTML to plain text conversion
- ✅ Email format validation
- ✅ Asynchronous sending (synchronous wait)

**Error Handling**:

| Error Code | Description |
|-----------|-------------|
| 400 | Email subject cannot be empty |
| 400 | Recipient list cannot be empty |
| 400 | Invalid email address |
| 500 | Email sending failed |

---

### V. Other Endpoints

#### 5.1 Health Check

**Endpoint**: `GET /ping`

**Authentication**: ❌ Not Required

**Function**: Check if server is running

**Request Example**:

```bash
curl -X GET http://localhost:8080/ping
```

**Response Example**:

```json
{
  "message": "pong"
}
```

---

#### 5.2 Get Homepage

**Endpoint**: `GET /`

**Authentication**: ❌ Not Required

**Function**: Return homepage HTML

**Request Example**:

```bash
curl -X GET http://localhost:8080/
```

---

#### 5.3 Swagger API Documentation

**Endpoint**: `GET /swagger/index.html`

**Authentication**: ❌ Not Required

**Function**: Interactive API documentation

**Access**: Open `http://localhost:8080/swagger/index.html` in browser

---

## 📊 Response Format Standards

### Success Response

```json
{
  "code": 200,
  "data": {},  // Response data
  "message": "success"
}
```

### Error Response

```json
{
  "code": 400,  // Or other error code
  "message": "Error description"
}
```

### HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Request parameter error |
| 401 | Unauthorized (missing or invalid Token) |
| 403 | Access forbidden |
| 404 | Resource not found |
| 429 | Request too frequent (rate limit) |
| 500 | Server internal error |

---

## 🔒 Security Notes

### Password Security

- ✅ Password encrypted with MD5 + random salt
- ✅ Salt stored separately in user table
- ✅ Password recovery not supported (keep your password safe)

### Token Security

- ✅ JWT Token signed with HS256 algorithm
- ✅ Valid for 7 days
- ✅ Expired Token requires re-login

### Request Limitations

- ✅ IP Rate Limiting: Maximum 5 requests per 10 seconds
- ✅ Exceeding limit returns 429 status code
- ✅ Automatic recovery based on IP

### Data Validation

- ✅ Phone: Must be 11 digits with format validation
- ✅ Email: Regex format validation
- ✅ Username: Cannot be empty
- ✅ Password: Cannot be empty

---

## 🧪 Testing Workflow

### 1. Registration Test

```bash
# Register user
curl -X POST http://localhost:8080/register \
  -F "name=testuser" \
  -F "password=123456" \
  -F "repeat_password=123456" \
  -F "phone=13800138000" \
  -F "email=test@example.com"
```

### 2. Login Test

```bash
# Login to get Token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'
```

### 3. Get User List

```bash
# Call protected endpoint with Token
curl -X GET http://localhost:8080/v1/user/list \
  -H "Authorization: <your_token>"
```

### 4. Add Friend

```bash
# Requires at least two users
curl -X POST http://localhost:8080/v1/relation/add \
  -H "Authorization: <your_token>" \
  -F "userId=1" \
  -F "targetName=2"
```

### 5. Send Email

```bash
# Send email notification
curl -X POST http://localhost:8080/v1/api/webhook \
  -H "Authorization: <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Test Email",
    "content":"This is a test email",
    "receivers":["test@example.com"]
  }'
```

---

## 📞 Frequently Asked Questions

### Q1: What to do if Token expires?

A: Token has 7-day validity. After expiration, re-login is required. Consider implementing Token refresh logic on client side.

### Q2: What if I forget my password?

A: Password recovery is not supported. Please keep your password safe.

### Q3: How to add a friend?

A: Use `/v1/relation/add` endpoint. You can add by user ID or username.

### Q4: What if email sending fails?

A: Check if email address format is correct and SMTP configuration is correct.

### Q5: Getting 429 error frequently?

A: Your IP exceeded rate limit (5 requests per 10 seconds). Please retry later.

---

## 📝 Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2024-01-01 | Initial release |

---

## 📄 Related Resources

- [Project Structure Documentation](./PROJECT_STRUCTURE_EN.md)
- [GitHub Repository](https://github.com/eric-jxl/emotionalbeach)
- [Swagger UI](http://localhost:8080/swagger/index.html)

---

**Last Updated**: 2024

