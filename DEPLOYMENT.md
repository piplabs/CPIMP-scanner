# CPIMP Scanner - GCP Compute Engine Deployment Guide

## Prerequisites

1. **GCP Account** with billing enabled
2. **gcloud CLI** installed and authenticated
3. **Git repository** (GitHub, GitLab, etc.) with your scanner code

## Option 1: Quick Deploy via Git (Recommended)

### Step 1: Create Compute Engine Instance

```bash
# Create a new VM instance
gcloud compute instances create cpimp-scanner \
    --zone=us-central1-a \
    --machine-type=e2-standard-2 \
    --boot-disk-size=20GB \
    --boot-disk-type=pd-standard \
    --image-family=ubuntu-2004-lts \
    --image-project=ubuntu-os-cloud \
    --tags=http-server,https-server
```

### Step 2: SSH into the instance

```bash
gcloud compute ssh cpimp-scanner --zone=us-central1-a
```

### Step 3: Setup the environment

```bash
# Download and run setup script
wget https://raw.githubusercontent.com/YOUR-USERNAME/CPIMP_scanner/main/setup_vm.sh
chmod +x setup_vm.sh
./setup_vm.sh
```

### Step 4: Clone and run your scanner

```bash
# Clone your repository
git clone https://github.com/YOUR-USERNAME/CPIMP_scanner.git
cd CPIMP_scanner

# Install dependencies
go mod tidy

# Start the scanner
./run_scanner.sh
```

## Option 2: Deploy via SCP

### Step 1: Create the VM (same as above)

### Step 2: Upload your code

```bash
# From your local machine, upload the entire project
gcloud compute scp --recurse . cpimp-scanner:~/CPIMP_scanner --zone=us-central1-a
```

### Step 3: SSH and setup

```bash
# SSH into the instance
gcloud compute ssh cpimp-scanner --zone=us-central1-a

# Run setup
cd ~/CPIMP_scanner
chmod +x setup_vm.sh
./setup_vm.sh

# Install dependencies and run
go mod tidy
./run_scanner.sh
```

## Monitoring and Management

### View logs
```bash
# Real-time log monitoring
tail -f ~/CPIMP_scanner/scanner.log

# Or use watch for periodic updates
watch -n 30 'tail -20 ~/CPIMP_scanner/scanner.log'
```

### Screen session management
```bash
# Attach to running scanner
screen -r cpimp-scanner

# List all screen sessions
screen -ls

# Detach from screen (while inside): Ctrl+A, then D

# Kill the scanner
screen -S cpimp-scanner -X quit
```

### Download results
```bash
# From your local machine, download CSV results
gcloud compute scp cpimp-scanner:~/CPIMP_scanner/*.csv . --zone=us-central1-a

# Download logs
gcloud compute scp cpimp-scanner:~/CPIMP_scanner/scanner.log . --zone=us-central1-a
```

## VM Management

### Stop the VM (save costs when not scanning)
```bash
gcloud compute instances stop cpimp-scanner --zone=us-central1-a
```

### Start the VM
```bash
gcloud compute instances start cpimp-scanner --zone=us-central1-a
```

### Delete the VM (when completely done)
```bash
gcloud compute instances delete cpimp-scanner --zone=us-central1-a
```

## Cost Optimization

- **e2-standard-2**: ~$50/month if running 24/7
- **Stop when not needed**: Only pay for storage (~$2/month for 20GB)
- **Preemptible instances**: Save 60-90% cost (can be interrupted)

```bash
# Create preemptible instance (much cheaper)
gcloud compute instances create cpimp-scanner-preemptible \
    --zone=us-central1-a \
    --machine-type=e2-standard-2 \
    --boot-disk-size=20GB \
    --preemptible \
    --image-family=ubuntu-2004-lts \
    --image-project=ubuntu-os-cloud
```

## Notifications

### Setup email alerts for completion
Add this to your scanner's config or modify the Go code to send emails when scanning completes or encounters errors using services like SendGrid or Gmail SMTP.

### Monitor via GCP Console
- Navigate to Compute Engine in GCP Console
- Click on your instance to view CPU/Memory usage
- Set up alerting policies for high resource usage

## Troubleshooting

1. **Out of memory**: Increase VM memory or reduce chunk size in scanner
2. **Network issues**: Check firewall rules and API rate limits
3. **Go build errors**: Ensure Go version compatibility (script installs Go 1.21)
4. **Permission errors**: Use `sudo` for system-level operations

## Security Notes

- VM has public IP by default
- Consider VPC and firewall rules for production
- Use service accounts with minimal permissions
- Regularly update the OS: `sudo apt update && sudo apt upgrade` 