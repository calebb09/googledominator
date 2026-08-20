# GoogleDominator Backend API (Golang + PostgreSQL Prisma)

Production-grade backend API for **GoogleDominator**, engineered in **Golang** using **PostgreSQL** with **Prisma ORM**, **Stripe Payment** checkout & webhooks, **Built for Taxes** forms catalog, dynamic **Pricing Plans**, and **Multi-step Onboarding** form handling.

---

## 🛠️ Tech Stack & Features

- **Language**: Go 1.22+
- **Framework**: [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database & ORM**: PostgreSQL with [Prisma Client Go](https://github.com/steebchen/prisma-client-go)
- **Payment Processing**: [Stripe Go SDK](https://github.com/stripe/stripe-go) (Checkout Sessions & Webhook events)
- **API Documentation**: Pre-built Postman v2.1.0 Collection included

---

## 🚀 Getting Started

### 1. Prerequisites

Make sure you have the following installed on your system:
- **Go** (v1.22 or higher) - `go version`
- **Node.js** (v18 or higher) - `node -v` (used for running Prisma CLI)
- **PostgreSQL** (Optional if running fallback mock mode)

---

### 2. Environment Configuration

Copy `.env.example` to create your local `.env` file:

```bash
cp .env.example .env
```

Default `.env` configuration:

```env
PORT=8080
ENV=development
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/googledominator?schema=public"
FRONTEND_URL="https://googledominator.co"
STRIPE_SECRET_KEY="sk_test_51MockStripeSecretKeyGoogleDominatorKey"
STRIPE_WEBHOOK_SECRET="whsec_mock_stripe_webhook_secret_key"
```

---

### 3. Install Dependencies & Generate Prisma Client

1. **Download Go Modules**:
   ```bash
   go mod tidy
   ```

2. **Generate Prisma Client for Go**:
   ```bash
   go run github.com/steebchen/prisma-client-go generate
   ```

3. **Push Schema to PostgreSQL Database** *(when PostgreSQL is running)*:
   ```bash
   npx prisma@6.19.0 db push --skip-generate
   ```

---

### 4. Running the Application

#### Option A: Run in Development Mode
```bash
go run cmd/api/main.go
```

#### Option B: Compile & Run Production Binary
```bash
# Build the binary
go build -o server cmd/api/main.go

# Run the binary
./server
```

The server will start listening on **`http://localhost:8080`**.

---

## 📌 API Endpoints Overview

| Category | Method | Endpoint | Description |
| :--- | :---: | :--- | :--- |
| **System** | `GET` | `/health` | Server health & DB connection status |
| **Pricing** | `GET` | `/api/v1/pricing` | List all pricing plans (Monthly & Annual) |
| | `GET` | `/api/v1/pricing/:id` | Get pricing plan by ID or slug |
| | `POST` | `/api/v1/pricing` | Create a new pricing tier (Admin) |
| **Built for Taxes** | `GET` | `/api/v1/taxes/services` | List tax packages & services |
| | `GET` | `/api/v1/taxes/services/:id` | Get tax service details |
| | `GET` | `/api/v1/taxes/forms` | List tax return forms (1040, 1099, 1120, W-2) |
| | `POST` | `/api/v1/taxes/services` | Create tax service offering (Admin) |
| **Onboarding** | `POST` | `/api/v1/onboarding` | Submit multi-step onboarding form |
| | `GET` | `/api/v1/onboarding` | List onboarding submissions |
| | `GET` | `/api/v1/onboarding/:id` | Get specific submission by ID |
| **Webinar** | `POST` | `/api/v1/webinar/register` | Register for webinar (Public) |
| | `GET` | `/api/v1/webinar` | List webinar registrations (Admin) |
| | `GET` | `/api/v1/webinar/:id` | Get webinar registration by ID (Admin) |
| | `PUT` | `/api/v1/webinar/:id` | Update webinar registration (Admin) |
| | `DELETE` | `/api/v1/webinar/:id` | Delete webinar registration (Admin) |
| **Stripe** | `POST` | `/api/v1/stripe/checkout-session` | Create Stripe Checkout Session |
| | `GET` | `/api/v1/stripe/session/:session_id` | Get & verify Stripe Checkout Session status |
| | `POST` | `/api/v1/stripe/webhook` | Process Stripe webhook events |
| | `GET` | `/api/v1/stripe/orders` | List purchase orders recorded in DB (Admin) |
| **Templates** | `POST` | `/api/v1/templates` | Create website template (Image upload form-data or JSON) |
| | `GET` | `/api/v1/templates` | List website templates |
| | `GET` | `/api/v1/templates/:id` | Get template by ID |
| | `PUT` | `/api/v1/templates/:id` | Update template |
| | `DELETE` | `/api/v1/templates/:id` | Delete template |

---

## 📬 Postman Collection

Import the included Postman collection into Postman to quickly test all endpoints:

📁 File Path: [`postman/GoogleDominator_API.postman_collection.json`](postman/GoogleDominator_API.postman_collection.json)

**How to Import into Postman**:
1. Open Postman.
2. Click **Import** in the top left.
3. Choose file: `postman/GoogleDominator_API.postman_collection.json`.
4. The environment variable `{{base_url}}` defaults to `http://localhost:8080`.

---

## 🧪 Testing the API (`curl` Examples)

### Health Check
```bash
curl -s http://localhost:8080/health
```

### Fetch Pricing Plans
```bash
curl -s http://localhost:8080/api/v1/pricing
```

### Webinar Registration
```bash
curl -X POST http://localhost:8080/api/v1/webinar/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1-555-234-5678",
    "registration_type": "tax professional",
    "additional_info": "Interested in CPA lead generation tactics.",
    "agreed": true
  }'
```

### Submit Onboarding Form
```bash
curl -X POST http://localhost:8080/api/v1/onboarding \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Apex Tax Practice",
    "contact_name": "Jane Doe, CPA",
    "phone": "+1-555-019-2831",
    "email": "jane@apextax.com",
    "has_existing_website": true,
    "website_url": "https://apextax.com",
    "has_google_business_profile": true,
    "primary_category": "Tax Preparation & CPA",
    "services_offered": "Form 1040, Form 1120S, Bookkeeping",
    "target_locations": "Dallas TX",
    "keywords": "tax prep near me, CPA Dallas",
    "visit_model": "Both In-Person & Virtual",
    "consent_transactional": true,
    "consent_marketing": true
  }'
```

### Create Stripe Checkout Session
```bash
curl -X POST http://localhost:8080/api/v1/stripe/checkout-session \
  -H "Content-Type: application/json" \
  -d '{
    "plan_slug": "new-website",
    "billing_cycle": "monthly",
    "email": "jane@apextax.com"
  }'
```
