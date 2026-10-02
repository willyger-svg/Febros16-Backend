# FEBROS16 - API Overview

This document serves as the primary contract between the FEBROS16 Backend and Frontend. 

## Base URL
- Development: `http://localhost:8080`
- Production: `https://febros16-backend.onrender.com` (or your configured Render URL)

## Authentication & Protected Middleware
Most endpoints require authentication. We use a **JWT (JSON Web Token)** approach.

**How the Middleware Works:**
1. The frontend must send the JWT in the HTTP headers:
   `Authorization: Bearer <your_jwt_token>`
2. The `RequireAuth` middleware intercepts the request.
3. It verifies the signature and expiration of the token using the `JWT_SECRET`.
4. If valid, it extracts the `user_id` and `role` and injects them into the Go `context (ctx)`.
5. The downstream handler (e.g., `GetMyProfile`) reads the `user_id` from the context to fetch user data securely without trusting client input.

## Standard Response Format
All successful responses return a JSON object with `success: true`.
```json
{
  "success": true,
  "message": "Optional success message",
  "data": { ... } // Contains the requested payload
}
```

## Standard Error Format
All errors return an appropriate HTTP status code (400, 401, 403, 404, 500) and `success: false`.
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE_STRING",
    "message": "Human readable error description"
  }
}
```

## Current Endpoints (Phase 2 - MVP)

### 1. Registration
- **Method:** `POST`
- **Path:** `/api/v1/auth/register`
- **Body:**
  ```json
  {
    "full_name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "success": true,
    "message": "Usajili umefanikiwa kikamilifu!"
  }
  ```

### 2. Login
- **Method:** `POST`
- **Path:** `/api/v1/auth/login`
- **Body:**
  ```json
  {
    "email": "john@example.com",
    "password": "securepassword123"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Umeingia kikamilifu!",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR..."
    }
  }
  ```

### 3. Get My Profile (Protected)
- **Method:** `GET`
- **Path:** `/api/v1/users/me`
- **Headers:** `Authorization: Bearer <token>`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "data": {
      "user": {
        "id": "uuid-here",
        "full_name": "John Doe",
        "email": "john@example.com",
        "role": "user",
        "bio": "Mtumiaji mpya wa FEBROS16",
        "profile_picture_url": "",
        "created_at": "2026-10-02T15:00:00Z",
        "updated_at": "2026-10-02T15:00:00Z"
      }
    }
  }
  ```
