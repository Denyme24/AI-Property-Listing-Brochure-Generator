# AI Property Listing Brochure Generator

An intelligent property listing brochure generator that leverages AI to create professional, customized property brochures in PDF format. The application uses OpenAI's GPT models to generate compelling property descriptions and generates beautifully formatted PDF brochures in both English and Arabic languages.

## 📹 Demo Videos

Check out our comprehensive demo videos to see the application in action:

**[View Demo Videos on Google Drive](https://drive.google.com/drive/folders/1xU1yfY0JPTt0ru6ibScIjkCbGYIMRylq)**

## Features

- **AI-Powered Content Generation**: Automatically generates engaging property descriptions using OpenAI GPT in both English and Arabic
- **Professional PDF Brochures**: Creates high-quality, professionally formatted PDF brochures with images in both English and Arabic languages
- **Cloud Storage**: Automatically uploads generated brochures and images to AWS S3
- **Modern UI**: Responsive web interface built with Next.js and React
- **Scalable Architecture**: Containerized application deployed on Kubernetes
- **Comprehensive Monitoring**: Real-time metrics, dashboards, and alerting with Prometheus, Grafana, and Slack integration

## Technology Stack

### Frontend
- **Framework**: Next.js 15.5.6 (with Turbopack)
- **Language**: TypeScript 5
- **UI Library**: React 19.1.0
- **Styling**: Tailwind CSS 4
- **UI Components**: shadcn/ui with Radix UI primitives
- **Form Management**: React Hook Form with Zod validation
- **Icons**: Lucide React
- **Theme**: next-themes for dark/light mode support

### Backend
- **Language**: Go 1.21
- **Web Framework**: Fiber v2
- **Database**: MongoDB Atlas
- **AI Integration**: OpenAI API (GPT models)
- **Cloud Storage**: AWS S3
- **PDF Generation**: gofpdf
- **Hot Reload**: Air (development)

### Deployment & Infrastructure
- **Containerization**: Docker (multi-stage builds)
- **Orchestration**: Kubernetes (EKS)
- **Container Registry**: Amazon ECR
- **Cloud Provider**: AWS
- **Region**: eu-north-1

### Monitoring & Observability
- **Metrics Collection**: Prometheus
- **Visualization**: Grafana
- **Alerting**: Alertmanager with Slack integration
- **Monitoring Stack**: Prometheus Operator (kube-prometheus-stack)

### CI/CD
- **Platform**: GitHub Actions
- **Workflow**: Automated build and deployment on backend changes
- **Process**:
  - Automatic Docker image builds
  - Push to Amazon ECR
  - Rolling deployment to EKS cluster

## Architecture

The application follows a modern microservices architecture with integrated monitoring:

```
Frontend (Next.js) → Backend API (Go/Fiber) → Services
                                              ├── MongoDB (Property Data)
                                              ├── OpenAI (Content Generation)
                                              └── AWS S3 (PDF Storage)
                          ↓
                    Prometheus (Metrics) → Grafana (Dashboards)
                          ↓
                    Alertmanager → Slack Notifications
```

## Prerequisites

- **Frontend**:
  - Node.js 20+
  - npm or yarn

- **Backend**:
  - Go 1.21+
  - MongoDB
  - AWS Account (for S3)
  - OpenAI API Key

- **Deployment**:
  - Docker
  - kubectl
  - AWS CLI
  - Amazon EKS cluster
  - Helm 3 (for monitoring stack deployment)

## Installation

### Frontend Setup

```bash
cd frontend
npm install
```

### Backend Setup

```bash
cd backend
go mod download
```

## Environment Variables

### Backend Configuration

Create a `.env` file in the `backend` directory:

```env
# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=property_brochure

# AWS Credentials
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=eu-north-1
AWS_S3_BUCKET=your_bucket_name

# OpenAI
OPENAI_API_KEY=your_openai_api_key

# Server
PORT=8000
```

### Frontend Configuration

Create a `.env.local` file in the `frontend` directory:

```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

## Running the Application

### Development Mode

**Backend**:
```bash
cd backend
# With hot reload (using Air)
air

# Or standard Go
go run main.go
```

**Frontend**:
```bash
cd frontend
npm run dev
```

The frontend will be available at `http://localhost:3000` and the backend API at `http://localhost:8000`.

### Production Build

**Backend**:
```bash
cd backend
docker build -t property-brochure-backend .
docker run -p 8000:8000 --env-file .env property-brochure-backend
```

**Frontend**:
```bash
cd frontend
npm run build
npm start
```

## Deployment

The application is deployed on AWS EKS with automated CI/CD:

1. **Kubernetes Configuration**: Deployment manifests are in `backend/K8s/`
   - `backend-deployment.yaml`: Deployment configuration
   - `backend-service.yaml`: Service configuration
   - `backend-env-secret.yaml`: Environment secrets

2. **CI/CD Pipeline**: GitHub Actions workflow (`.github/workflows/ci-cd.yaml`)
   - Triggers on changes to `backend/**`
   - Builds Docker image
   - Pushes to Amazon ECR
   - Updates EKS deployment

3. **Required GitHub Secrets**:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_ACCOUNT_ID`

## Monitoring & Observability

The application includes a comprehensive monitoring stack for tracking metrics, visualizing data, and sending alerts.

### Components

- **Prometheus**: Collects and stores metrics from the Kubernetes cluster and applications
- **Grafana**: Provides beautiful dashboards for visualizing metrics and monitoring system health
- **Alertmanager**: Manages alerts and routes them to configured receivers (Slack)

### Setup

#### 1. Install Prometheus Stack

Deploy the kube-prometheus-stack using Helm with custom values:

```bash
# Add Prometheus Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Create monitoring namespace
kubectl create namespace monitoring

# Install Prometheus stack with custom configuration
helm install prometheus-stack prometheus-community/kube-prometheus-stack \
  -n monitoring \
  -f backend/K8s/prometheus-values.yaml
```

#### 2. Configure Slack Alerting

Create a Slack webhook URL ([Slack webhook creation guide](https://api.slack.com/messaging/webhooks)) and configure Alertmanager:

```bash
# Apply Slack secret configuration
kubectl apply -f backend/K8s/alertmanager-slack.yaml
```

**Note**: Update the Slack webhook URL in `alertmanager-slack.yaml` with your own webhook before applying.

#### 3. Access Monitoring Dashboards

**Grafana**:
```bash
# Access via NodePort (default: 32000)
kubectl get svc -n monitoring prometheus-stack-grafana
```

**⚠️ SECURITY WARNING**: The default credentials are `admin/admin`. **Change the password immediately** after first login to secure your Grafana instance. You can configure a custom password in `prometheus-values.yaml` before installation.

**Prometheus**:
```bash
# Access via NodePort (default: 32090)
kubectl get svc -n monitoring prometheus-stack-kube-prom-prometheus
```

### Features

- **Real-time Metrics**: Monitor CPU, memory, network, and application-specific metrics
- **Custom Dashboards**: Pre-configured Grafana dashboards for Kubernetes monitoring
- **Alerting Rules**: Automated alerts for critical issues (high CPU, memory, pod crashes)
- **Slack Notifications**: Instant alerts delivered to your Slack channel
- **Historical Data**: Query and analyze historical metrics for troubleshooting

## API Endpoints

The backend exposes the following main endpoints:

- `POST /api/property` - Submit property details and generate brochure
- Additional endpoints for property management

## Project Structure

```
.
├── frontend/                 # Next.js frontend application
│   ├── app/                 # Next.js app directory
│   ├── components/          # React components
│   ├── lib/                 # Utility functions
│   └── public/              # Static assets
│
├── backend/                 # Go backend application
│   ├── handlers/            # HTTP request handlers
│   ├── services/            # Business logic services
│   │   ├── mongodb.go      # Database operations
│   │   ├── openai.go       # AI content generation
│   │   ├── pdf.go          # PDF generation
│   │   └── s3.go           # Cloud storage
│   ├── models/             # Data models
│   ├── middleware/         # HTTP middleware
│   ├── config/             # Configuration management
│   ├── K8s/                # Kubernetes manifests
│   │   ├── backend-deployment.yaml      # Backend deployment config
│   │   ├── backend-service.yaml         # Backend service config
│   │   ├── backend-env-secret.yaml      # Environment secrets
│   │   ├── prometheus-values.yaml       # Prometheus & Grafana config
│   │   └── alertmanager-slack.yaml      # Slack alerting config
│   └── Dockerfile          # Container configuration
│
└── .github/
    └── workflows/
        └── ci-cd.yaml      # CI/CD pipeline
```
