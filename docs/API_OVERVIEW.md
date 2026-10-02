# FEBROS16 - API Overview

This document serves as the primary contract between the FEBROS16 Backend and Frontend. 

## Base URL
- Development: `http://localhost:8080`
- Production: `https://febros16-backend.onrender.com` (or your configured Render URL)

## Authentication
Most endpoints (except public ones like login/register) require a JWT token in the `Authorization` header.
Format: `Authorization: Bearer <your_jwt_token>`

## Response Format
All successful responses return a JSON object with a status code of `200` or `201`.
```json
{
  "message": "Success message or description",
  "data": { ... } // Optional: Contains the requested payload
}
```

## Error Format
All errors return an appropriate HTTP status code (e.g., 400, 401, 403, 404, 500) and a standardized JSON body:
```json
{
  "error": "Short error code or description"
}
```

## Current Endpoints (MVP)

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
    "message": "Umeingia kikamilifu!",
    "token": "eyJhbGciOiJIUzI1NiIsInR..."
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
        "created_at": "2026-10-02...",
        "profile": {
          "id": "uuid",
          "user_id": "uuid-here",
          "bio": "Mtumiaji mpya wa FEBROS16",
          "avatar_url": "",
          "website": ""
        }
      }
    }
  }
  ```
