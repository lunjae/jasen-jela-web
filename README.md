# Jasen Jela web katalog

Production-oriented full-stack catalog and administration application for coffin products. The visitor-facing experience is in Serbian Latin; source code and API contracts are in English. The application intentionally has no checkout or payment flow.

## Architecture

- `frontend/`: React 19, TypeScript, Vite, React Router, TanStack Query, React Hook Form, Zod and Firebase Authentication.
- `backend/`: Go `net/http` API, Firebase Admin token verification, Firestore repository and signed Cloudinary image uploads.
- `firebase/`: Firestore security rules plus local emulator configuration.
- Public reads and inquiry writes go through the backend. Admin endpoints verify the Firebase ID token and then independently require an `admins/{uid}` Firestore document.
- `USE_MEMORY_STORE=true` provides an ephemeral local repository and `X-Dev-Admin-UID: local-admin` development identity. It must never be enabled in production.

The backend uses consistent `{ "data": ..., "error": null }` and `{ "data": null, "error": ... }` envelopes, body limits, validation, request recovery, configurable CORS, timeouts, and graceful shutdown. Firestore is hidden behind a repository interface to keep handlers testable.

## Prerequisites

- Go 1.26 or compatible version from `backend/go.mod`
- Node.js 22+ and npm
- A Firebase project with Authentication and Firestore enabled
- A Cloudinary account for product image storage and delivery
- Firebase CLI (`npm install -g firebase-tools`) for rules/emulators

## Environment configuration

Copy examples and fill local values; do not commit the resulting files:

```sh
cp frontend/.env.example frontend/.env
cp backend/.env.example backend/.env
```

Vite Firebase values are public client configuration, not service credentials. The backend uses Application Default Credentials or the service-account file named by `GOOGLE_APPLICATION_CREDENTIALS`. Never place that JSON file in `frontend/`.

The `make dev-backend` command loads `backend/.env` before starting Go. Cloudinary credentials are required in both memory and Firestore modes because all image uploads use Cloudinary.

## Firebase setup

1. Create a Firebase web application and copy its client configuration to `frontend/.env`.
2. Enable Email/Password in Authentication.
3. Create Firestore.
4. Create/download a service-account key for the backend, store it outside Git, and set `GOOGLE_APPLICATION_CREDENTIALS`.
5. Set `FIREBASE_PROJECT_ID` and `USE_MEMORY_STORE=false` for the backend.
6. From `firebase/`, select the project with `firebase use --add`, then deploy rules and indexes with `firebase deploy --only firestore:rules,firestore:indexes`.

## Cloudinary setup

Create a Cloudinary account and copy the cloud name, API key, and API secret from its dashboard into the backend environment as `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, and `CLOUDINARY_API_SECRET`. The secret must only be available to the backend. Uploads and deletions are signed server-side; no Cloudinary credentials are exposed to the browser.

### Creating the first administrator

Create the user in Firebase Authentication. Copy its UID and create this document from the Firebase console or a trusted Admin SDK script:

```text
admins/{firebase-auth-uid}
  enabled: true
```

The document must be created through a trusted environment because security rules deny client writes. The frontend never supplies or decides the role.

## Local development

Install dependencies once:

```sh
cd frontend && npm install
cd ../backend && go mod download
```

Run backend and frontend together:

```sh
make dev
```

Alternatively, run them in separate terminals:

```sh
make dev-backend
make dev-frontend
```

Open `http://localhost:5173`. In development memory mode, protected API calls use the local admin identity, while image uploads still use Cloudinary. Firebase login is still the normal UI path; configure Firebase to exercise the complete authentication flow.

Start Firebase emulators with `make emulators`. When using them, configure the standard emulator environment variables and connect the frontend SDK to the emulators before relying on emulator-only auth; the default frontend configuration targets the selected Firebase project.

## API overview

Public:

- `GET /api/health`
- `GET /api/products` — `search`, `category`, `material`, `color`, `minPrice`, `maxPrice`, `sort`, `page`, `pageSize`
- `GET /api/products/{slug}`
- `GET /api/categories`
- `POST /api/inquiries`

Admin routes under `/api/admin` manage products, image uploads/removals, categories, inquiries, and inquiry status. They require `Authorization: Bearer <Firebase ID token>`. See `backend/cmd/api/main.go` for the complete route table.

Image uploads accept up to eight files in the repeated `multipart/form-data` field `images` (the singular `image` field remains accepted), plus optional `alt` and `isPrimary` fields. The backend permits verified JPEG, PNG, or WebP content up to 10 MB per image. Production uploads use generated Cloudinary public IDs and store both the secure delivery URL and public ID. Primary selection and complete image ordering are managed through dedicated authenticated API endpoints.

Images are uploaded under `jasen-jela/products/{productId}/`. Firestore stores only `url`, `publicId`, `alt`, `order`, and `isPrimary`; image bytes and Cloudinary credentials are never written to Firestore. Legacy Firestore image entries containing `storagePath` are read and returned as `publicId`, then migrated on the next product write.

### Firestore collections

- `admins/{firebaseUid}`: `enabled` (must be explicitly `true`).
- `categories/{categoryId}`: `name`, `slug`, `description`, `published`, `createdAt`, `updatedAt`.
- `products/{productId}`: product text fields, `categoryId`, `material`, `color`, optional `dimensions` and `price`, `currency`, `featured`, `published`, normalized `images`, `createdAt`, `updatedAt`.
- `inquiries/{inquiryId}`: optional `productId`, `fullName`, `email`, `phone`, `message`, `status`, `createdAt`, `updatedAt`.

The React application does not initialize Firestore. All reads and writes use the Go API; Firebase client code is used for Authentication only. Firestore client rules deny all direct reads and writes. The current repository performs flexible filtering and pagination after collection reads, so no composite indexes are required yet; `firebase/firestore.indexes.json` intentionally contains no indexes.

## Seed data

Run `make seed` to insert three Serbian categories and four products without fake image uploads. It refuses to run when products or categories already exist. For an intentional additive development run use `make seed SEED_FLAGS=--allow-non-empty`. Production additionally requires `SEED_FLAGS="--allow-non-empty --allow-production"`.

## Quality and build commands

```sh
make format  # gofmt and Prettier
make lint    # go vet and oxlint
make test    # Go and Vitest suites
make build   # backend binary and frontend production bundle
```

Individual commands are available through `frontend/package.json` and standard Go tooling. Build output is `backend/bin/api` and `frontend/dist/`.

## Deployment notes

- Deploy the backend on a service with HTTPS, Firebase credentials through workload identity/secret management, and `USE_MEMORY_STORE=false`.
- Set `FRONTEND_ORIGIN` to the exact HTTPS frontend origin. Production configuration rejects `*`.
- Build the frontend with its production API URL and Firebase web configuration. Configure the host to fall back to `index.html` for React routes.
- Deploy Firebase rules before opening the application to users.
- Replace the example company phone, email, address, and existing starter hero artwork with final approved business content.
- Product images are delivered through Cloudinary HTTPS URLs and removed through signed backend requests.
- Add rate limiting and abuse protection at the reverse proxy/API gateway for the public inquiry endpoint.

## Security considerations

Secrets and environment files are ignored. Admin authorization is checked server-side against Firestore after token verification. Validation is duplicated at the UI boundary and API boundary. Logs contain method/path/timing but not tokens, inquiry bodies, or credentials. Public Firestore inquiry creation is denied because submission is mediated by the backend.
