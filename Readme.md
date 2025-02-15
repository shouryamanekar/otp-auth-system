# OTP Authentication System (Golang + Redis + PostgreSQL)

## Overview
The OTP Authentication System is a secure, scalable authentication service built with Golang, Redis, and PostgreSQL. It supports:
- User Registration (without OTP verification)
- Login with OTP (sent via Fast2SMS)
- JWT-Based Authentication
- Multi-Device Support
- Session Management (Logout, Logout-All)
- Token Blacklisting (Prevent reuse of logged-out tokens)
- OTP Rate Limiting (Prevents SMS spam)
- Automatic Expired Token Cleanup (Optimized Redis memory usage)
- API Documentation with Swagger

## Tech Stack
- Backend: Golang (Gin Framework)
- Database: PostgreSQL
- Cache: Redis (JWT Blacklisting, OTP Storage)
- SMS Gateway: Fast2SMS
- Containerization: Docker
- Deployment: Heroku

---

## Installation & Setup

### 1. Clone the Repository
```sh
git clone https://github.com/your-repo/otp-auth-system.git
cd otp-auth-system
```

### 2. Setup Environment Variables
Create a `.env` file and configure your environment:
```sh
PORT=8080
DATABASE_URL=postgres://user:password@host:port/dbname
REDIS_URL=rediss://:password@host:port
FAST2SMS_API_KEY=your_fast2sms_api_key
JWT_SECRET=your_jwt_secret
```

### 3. Install Dependencies
```sh
go mod tidy
```

### 4. Run the Application
```sh
go run main.go
```
The server should be running on `http://localhost:8080`.

---

## Deployment (Heroku)

### 1. Login to Heroku
```sh
heroku login
```

### 2. Deploy the App
```sh
git add .
git commit -m "Deploy OTP Authentication System"
git push heroku main
```

### 3. Set Environment Variables
```sh
heroku config:set DATABASE_URL=postgres://user:password@host:port/dbname
heroku config:set REDIS_URL=rediss://:password@host:port
heroku config:set FAST2SMS_API_KEY=your_fast2sms_api_key
heroku config:set JWT_SECRET=your_jwt_secret
```

### 4. Restart the App
```sh
heroku restart --app otp-auth-system
```

---

## API Endpoints
### Authentication
| Method | Endpoint         | Description |
|--------|-----------------|-------------|
| `POST` | `/register`      | Register a new user (No OTP required) |
| `POST` | `/login`         | Request OTP for login |
| `POST` | `/verify`        | Verify OTP and issue JWT |
| `POST` | `/resend-otp`    | Resend OTP (Rate-limited) |

### User Management
| Method  | Endpoint  | Description |
|---------|-----------|-------------|
| `GET`   | `/user`   | Get current user details |
| `GET`   | `/user/devices` | Get all registered devices |
| `DELETE`| `/device` | Remove a specific device |
| `DELETE`| `/devices/all` | Remove all devices |

### Logout
| Method  | Endpoint        | Description |
|---------|----------------|-------------|
| `POST`  | `/logout`      | Logout from the current device |
| `POST`  | `/logout/all`  | Logout from all devices |

### API Documentation
Swagger UI is available at:
```
https://otp-auth-system.herokuapp.com/swagger/index.html
```

---

## How It Works

### 1. User Registration
- Users register with their mobile number.
- OTP is not sent during registration.

### 2. User Login
- Users request an OTP via `/login`.
- OTP is stored in Redis for 5 minutes.
- After entering OTP via `/verify`, JWT is issued.

### 3. Multi-Device Support
- Users can log in on multiple devices.
- Each device has its own JWT.
- The latest device is always kept active on `/logout/all`.

### 4. Token Blacklisting
- When users log out, their JWT token is blacklisted in Redis.
- Blacklisted tokens cannot be used even if valid.

### 5. Redis Memory Optimization
- Expired tokens are automatically cleaned up.
- Rate-limiting prevents OTP spam via Fast2SMS.

---

## Security Features
- JWT Authentication (Tokens expire in 24 hours)
- Token Blacklisting (Prevents reuse after logout)
- OTP Rate Limiting (Max 3 OTPs per hour per user)
- Multi-Device Management (Users can see/remove logged-in devices)
