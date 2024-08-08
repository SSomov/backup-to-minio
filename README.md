# Backup Script with Minio and Docker Volume Support

This project provides a flexible and configurable backup solution that supports both local directories and Docker volumes. The backups are stored on Minio using the Minio client (`mc`).

## Features

- **Backup Local Directories**: Easily back up folders to Minio.
- **Backup Docker Volumes**: Export Docker volumes to a file and store it on Minio.
- **Configuration via YAML**: Define what needs to be backed up in a simple `backup.yml` file.
- **Environment Variables**: Minio credentials and endpoint are managed via a `.env` file.

## Getting Started

### Prerequisites

- Docker
- Minio (or S3-compatible storage)

### Configuration

1. **Environment Variables**

   Create a `.env` file in the root directory:

   ```bash
   MINIO_ENDPOINT=http://minio-endpoint:9000
   MINIO_ACCESS_KEY=minio_access_key
   MINIO_SECRET_KEY=minio_secret_key

2. **Backup Configuration**

    Define the resources you want to back up in backup.yml
