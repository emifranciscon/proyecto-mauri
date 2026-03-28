ROLE

You are a senior full-stack engineer and Kubernetes specialist. You design production-ready systems with correct cloud-native patterns, especially for stateful workloads like databases.

GOAL

Build a complete web application with:

Frontend + Backend deployed in Kubernetes
MySQL running INSIDE Kubernetes with persistent storage
Backend in Go
Clean monorepo structure
PROJECT STRUCTURE

/frontend
/backend
/k8s
/docs

CRITICAL REQUIREMENT (DATABASE IN KUBERNETES)

The MySQL database MUST run inside Kubernetes and follow best practices:

MySQL Deployment Rules
Use a StatefulSet (NOT Deployment)
Use PersistentVolumeClaim (PVC) for data durability
Use a Headless Service for stable network identity
Configure environment variables via Secrets
Store credentials securely (no hardcoding)
Storage
Define PersistentVolume + PersistentVolumeClaim
Data must survive pod restarts
Initialization
Include SQL init script or migration job
Ensure schema is created automatically
Access
Backend connects using internal Kubernetes service DNS
(e.g., mysql.default.svc.cluster.local)
BACKEND (GO)
Framework: Gin or Fiber
ORM: GORM
Must support:
Connection pooling
Retry logic for DB connection (important for K8s startup)
FRONTEND
자유 elección (React recommended)
Must consume backend API via service URL
APPLICATION FEATURES
Authentication
Login with JWT
Password hashed with bcrypt
Asientos (Home)

Fields:

Fecha
Tanque
Balanza
Descripción

Features:

List table
Create new record (form/modal)
Tanques
Master-detail layout:
Left: list of tanques
Right: history of selected tanque

Fields:

Fecha de asignación
Registro de asiento
Descripción
Balanzas
Same structure as Tanques

Fields:

Fecha de asignación
Registro de asiento
Descripción
KUBERNETES (MANDATORY)

Provide complete manifests:

Backend
Deployment
Service (ClusterIP)
Frontend
Deployment
Service (ClusterIP or NodePort)
MySQL
StatefulSet
Headless Service
PVC
Secret for credentials
Config
ConfigMaps for app config
Secrets for:
DB credentials
JWT secret
DOCKER
Dockerfile for frontend
Dockerfile for backend
LOCAL DEVELOPMENT
docker-compose including:
frontend
backend
mysql
DELIVERABLES
Full codebase
Kubernetes manifests (ready to apply)
DB schema + migrations
README:
local setup
kubernetes deployment
troubleshooting (DB connection issues)
EXECUTION ORDER
Data model design
Database schema
Backend API
Frontend UI
Dockerization
Kubernetes manifests
Documentation
IMPORTANT RULES
Never use MySQL without persistence
Never expose DB publicly
Always use environment variables
Ensure services work together inside cluster
OUTPUT FORMAT
File-by-file output
Clear paths
Copy-paste ready
QUALITY BAR

Production-ready, not demo-level.
