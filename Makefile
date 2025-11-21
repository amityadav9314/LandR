.PHONY: help start-backend start-frontend start stop clean

# Default target
help:
	@echo "Learn and Revise - Development Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make start          - Start both backend and frontend servers"
	@echo "  make start-backend  - Start only the backend server"
	@echo "  make start-frontend - Start only the frontend server"
	@echo "  make stop           - Stop all running servers"
	@echo "  make clean          - Clean build artifacts"
	@echo ""

# Start both servers (run in separate terminals)
start:
	@echo "Starting backend and frontend servers..."
	@echo "Backend will run on :50051 (gRPC) and :8080 (gRPC-Web)"
	@echo "Frontend will run on :8081 (Metro Bundler)"
	@echo ""
	@echo "Run these commands in separate terminals:"
	@echo "  Terminal 1: make start-backend"
	@echo "  Terminal 2: make start-frontend"

# Start backend server
start-backend:
	@echo "Starting backend server..."
	@echo "Note: Configure backend/.env file with your API keys"
	@echo ""
	cd backend && go run cmd/server/main.go

# Start frontend server
start-frontend:
	@echo "Starting frontend Expo server..."
	cd frontend && cmd /c npm start

# Stop servers (Windows)
stop:
	@echo "Stopping servers on ports 8080, 8081, and 50051..."
	@for /f "tokens=5" %a in ('netstat -ano ^| findstr ":8080 :8081 :50051"') do @taskkill /F /PID %a 2>nul || echo No process found

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cd backend && del /Q bin\* 2>nul || echo No backend artifacts to clean
	cd frontend && rmdir /S /Q node_modules\.cache 2>nul || echo No frontend cache to clean
	@echo "Clean complete!"
