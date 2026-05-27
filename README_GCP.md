# 🌐 Operating TodoApp in Google Cloud Platform (GCP)

This guide provides detailed instructions on running, building locally, manually testing cloud builds, and monitoring the **TodoApp** application in your GCP infrastructure.

The application consists of two main services running in **Google Cloud Run**:
1. **backend-api** (Go service, port 8080) which connects to a private **Cloud SQL (PostgreSQL)** database instance via a Serverless VPC Access Connector.
2. **frontend-app** (React + Nginx, port 80) which interacts with the backend using the backend's public HTTPS URL.

---

## 🚀 1. Local Launch and Testing

Before deploying to the cloud, you can test both components locally.

### A. Running the Database (PostgreSQL) Locally via Docker
To run the backend locally, you need a PostgreSQL database instance:
```bash
docker run --name local-postgres \
  -e POSTGRES_USER=dbuser \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=todoapp \
  -p 5432:5432 \
  -d postgres:15-alpine
```

### B. Running the Backend (Go) Locally
Navigate to the backend directory and start the server:
```bash
cd backend
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=dbuser
export DB_PASSWORD=password
export DB_NAME=todoapp
export listenPort=8080

go run main.go metrics.go
```
*Verify the backend is up and running:*
```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/todos
```

### C. Running the Frontend (React) Locally
The frontend requires the API URL of your local backend. Navigate to the frontend directory and start the development server:
```bash
cd frontend
export REACT_APP_API_URL=http://localhost:8080
npm install
npm start
```
The application will open automatically at `http://localhost:3000`.

---

## 🛠️ 2. Building Docker Images Locally

To verify your Dockerfile configurations build successfully on your local machine:

### Building the Backend:
```bash
cd backend
docker build -t europe-central2-docker.pkg.dev/test-122133/app-repo/backend-api:local .
```

### Building the Frontend (injecting the cloud backend API URL):
```bash
cd frontend
docker build \
  --build-arg REACT_APP_API_URL="https://backend-api-${PROJECT_ID}$.europe-central2.run.app" \
  -t europe-central2-docker.pkg.dev/${PROJECT_NAME}/app-repo/frontend-app:local .
```

---

## ☁️ 3. Manual Build Trigger and Verification via Google Cloud Build

You can test and run your cloud pipeline **directly from your local terminal using Cloud Build**, without doing a git push to GitHub. This is ideal for quickly validating your `cloudbuild.yaml` structure.

### Running a Manual Cloud Build via gcloud CLI:
Ensure you are authenticated in the gcloud SDK and are in the root directory of your workspace 
```bash
gcloud auth login
gcloud config set project ${PROJECT_NAME}

gcloud builds submit --config=cloudbuild.yaml \
  --substitutions=_PREFIX="${PROJECT_NAME}",_REGION="${REGION}",_REPOSITORY="app-repo"
```

This command will:
1. Package your local files (respecting `.gitignore` rules).
2. Upload the source archive to Google Cloud Build.
3. Trigger compilation, unit testing, building Docker images, pushing them to the Artifact Registry, and deploying/updating the services on Cloud Run.

---

## 🤖 4. Automatic CI/CD Pipelines (GitHub Triggers)

Once the infrastructure is provisioned via Terraform, deployments occur automatically:

1. **GitHub Trigger (Cloud Build)**:
   On every `git push` to the `main` branch, the `todoapp-deploy-trigger` trigger reads your [cloudbuild.yaml](cloudbuild.yaml) file and executes the cloud build and deployment revision rollout.
2. **GitHub Actions**:
   An alternative automated workflow is configured in [.github/workflows/google-ci.yml](.github/workflows/google-ci.yml), which builds and rolls out updates using a secure repository secret named `GCP_SA_KEY` containing your service account JSON key.

---

## 🔎 5. Verification and Monitoring via Console (gcloud & URL)

Once deployed successfully, you can verify your service status and configuration using CLI commands.

### A. List Cloud Run Services and URLs:
```bash
gcloud run services list --project=${PROJECT_NAME} --region=${REGION}
```
*The command output will list the exact public URL endpoints for both frontend and backend.*

### B. Verify Backend Health and Connectivity using `curl`:
Check the health status (which returns the correct CORS headers):
```bash
curl -i https://backend-api-${PROJECT_NUMBER}.${REGION}.run.app/health
```
Fetch the tasks list:
```bash
curl -i https://backend-api-${PROJECT_NUMBER}.${REGION}.run.app/todos
```

### C. Live Cloud Run Logs Monitoring:
Review backend logs to diagnose database connection status, CORS headers, or other internal errors:
```bash
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=backend-api" \
  --limit=20 --project=${PROJECT_NAME}
```
To stream live container logs (live tail):
```bash
gcloud beta run services logs tail backend-api --project=${PROJECT_NAME} --region=${REGION}
```
