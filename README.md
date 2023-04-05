# cloud-cleaner

## Overview

cloud-cleaner's job is very simple. It is to clean and optimize cloud resources.

## How to Use

This repository is a a backend server.

Latest Swagger Spec is available at https://github.com/jay-babu/cloud-cleaner/blob/main/oapi/todo.yml

## Supported Services
- CloudWatch Log Groups
  - Retention Policy

## Running Locally

- Assumes AWS Credentials from a profile. By default, it looks for a profile named `default`.
  - To override the profile, set an environment variable `PROFILE` to the desired profile.
- `./cloud-cleaner` if downloaded release file
- `go run main` if running from source
