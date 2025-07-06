# 200 Days of Code - Progress Log

This document tracks the completion of projects in the 200 Days of Code challenge.

---

## Completed Projects

### Real-time Chat Application
- Initial setup with Node.js, Express, Socket.IO.
- Implemented PostgreSQL for database.
- Added user registration and login with password hashing (bcrypt) and session management.
- Developed a modern, responsive UI using Bootstrap and Material Design principles for authentication and chat interfaces.
- Implemented basic private messaging functionality.
- Implemented basic group chat functionality.

**Current Roadblock:** Server error on sign-in due to PostgreSQL connection issues. Need to verify `db.js` credentials match Docker container setup and ensure the database exists within the container.

**Next Steps:**
1.  Resolve PostgreSQL connection issue.
2.  Refine UI/UX (e.g., better styling for active chats/groups, notifications).
3.  Implement persistent chat history (store messages in the database).
4.  Add more advanced chat features (e.g., typing indicators, read receipts).
5.  Work on deployment to a cloud platform.

### E-commerce Platform
**Approach:**
- **Frontend:** React (JavaScript)
- **Backend (API Gateway):** Node.js (Express)
- **Backend (Database Layer):** Python (FastAPI) communicating with PostgreSQL (containerized)

**Progress:**
- Created project directory.
- Created `README.md`.
- Set up Node.js backend (`backend-nodejs`): `package.json`, installed dependencies, created `server.js`.
- Set up Python backend (`backend-python`): `requirements.txt`, `Dockerfile`, `main.py`, `database.py`, `schemas.py`, `models.py`.

**Current Roadblock:** Unable to build Docker image for Python backend due to `docker buildx build` error, even after attempting to disable BuildKit. This indicates a potential issue with the Docker Desktop setup or environment configuration on the user's machine.

**Next Steps:**
1.  Resolve Docker build issue for Python backend.
2.  Build and run Python backend Docker container.
3.  Start Node.js backend.
4.  Set up React frontend.

