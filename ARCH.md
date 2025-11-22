# Learn and Revise (LandR) Architecture

## Overview
LandR is a SaaS application for learning and revision, allowing users to convert materials (text/links) into flashcards using AI.

## Technology Stack

### Backend
- **Language**: Go (Golang)
- **Framework**: gRPC (Google Protocol Buffers)
- **Database**: PostgreSQL (with `pgx` driver)
- **AI**: Groq API (LLM for flashcard generation)
- **Migrations**: `golang-migrate`

### Frontend
- **Framework**: React Native (Expo)
- **Language**: TypeScript
- **Communication**: gRPC-Web (via `nice-grpc`)
- **State Management**: Zustand (implied from `authStore.ts`)
- **Navigation**: React Navigation

## Database Schema

### Tables
1.  **`users`**
    *   `id` (UUID, PK)
    *   `email`, `name`, `google_id`, `picture`
    *   Stores user profile and authentication info.

2.  **`materials`**
    *   `id` (UUID, PK)
    *   `user_id` (FK -> `users.id`)
    *   `type` (TEXT/LINK), `content`, `title`
    *   Stores the source content for learning.

3.  **`flashcards`**
    *   `id` (UUID, PK)
    *   `material_id` (FK -> `materials.id`)
    *   `question`, `answer`
    *   `stage` (Spaced Repetition stage), `next_review_at`
    *   Stores generated flashcards and their review state.

4.  **`tags`**
    *   `id` (UUID, PK)
    *   `user_id` (FK -> `users.id`)
    *   `name`
    *   Stores user-defined tags for categorization.

5.  **`material_tags`**
    *   `material_id` (FK -> `materials.id`)
    *   `tag_id` (FK -> `tags.id`)
    *   Join table for Many-to-Many relationship between Materials and Tags.

### Relationships
-   **User -> Materials**: One-to-Many (Cascade Delete)
-   **Material -> Flashcards**: One-to-Many (Cascade Delete)
-   **User -> Tags**: One-to-Many
-   **Material <-> Tags**: Many-to-Many (via `material_tags`)

## Backend Architecture
Follows **Clean Architecture** principles:
1.  **Transport Layer (`internal/service`)**: gRPC handlers (`LearningService`). Handles request/response mapping.
2.  **Business Logic (`internal/core`)**: Core application logic (`LearningCore`). Orchestrates AI generation and DB operations.
3.  **Data Access (`internal/store`)**: Database implementations (`PostgresStore`). Executes SQL queries.
4.  **External Services (`internal/ai`)**: Clients for external APIs (`GroqClient`).

## Frontend Architecture
1.  **Screens**: UI pages (e.g., `HomeScreen`).
2.  **Components**: Reusable UI elements.
3.  **Stores**: Global state management (e.g., `authStore`).
4.  **API Client**: Generated gRPC-Web client for communicating with the backend.

## Data Flow
1.  **Add Material**:
    -   Frontend sends `AddMaterialRequest` (Content + Tags).
    -   Backend calls AI to generate Flashcards, Title, and Tags.
    -   Backend saves Material, Tags, and Flashcards to DB in a transaction-like manner.
    -   Returns created data to Frontend.

2.  **Review Flashcards**:
    -   Frontend requests `GetDueFlashcards`.
    -   Backend queries `flashcards` joined with `materials` and `tags`.
    -   Frontend displays grouped flashcards.
