# Stripe checkout deployed on GCP Cloud Run

This configuration contains a server in go deployed on GCP Cloud Run service that initiates a Stripe session according to configuration passed from Terraform configuration.  
It also optionally creates a mapping between a Cloud Run service and a custom domain.

## Building a container image

Use [skaffold](https://skaffold.dev) to build and push the image to GCR.

1. Go to [server/](server/) 
2. Set a tag for the container image using `RELEASE` environment variable

```bash
export RELEASE=0.1.0
```

3. Set a destination registry and repo for the image using skaffold's `SKAFFOLD_DEFAULT_REPO` environment variable

```bash
export SKAFFOLD_DEFAULT_REPO=gcr.io/YOUR-GCP-PROJECT
```

4. Build and push the image

```bash
skaffold build
```

## Deploying to Cloud Run with Terraform



1. Go to [terraform/](terraform/) 
2. Create the `terraform.tfvars` file from `terraform.tfvars.example` and set variables for the service, including the full name of the container image built previously.
3. Apply this configiration with terraform

```bash
terraform apply
```
