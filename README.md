# Subscription Invoicing System

## Problem

Managing a diverse array of customer contracts and ensuring timely invoicing is a challenge in subscription services. The system needs to handle subscriptions with varied durations and billing intervals efficiently to ensure accurate and timely invoicing.

## Approach

1. **Database Structuring**:
    - Designed a schema to organize customer contract information within the database.
    - Utilized GoLang with PostgreSQL for seamless data processing and retrieval.
    - Ensured the schema facilitates easy management and retrieval of contract details.

2. **Daily Scheduled Job Implementation**:
    - Implemented two daily cron jobs:
        - One for processing invoices billed on that day.
        - Another for retrying to send failed invoices.
    - Employed a sophisticated algorithm to identify customers due for invoicing based on contract start dates and billing frequencies.
    - Ensured accuracy and reliability of the cron job by handling exceptions and error scenarios effectively.

3. **Invoice Generation Workflow**:
    - Created a comprehensive workflow for generating, formatting, and delivering invoices:
        - Conducted internal calculations for invoice amounts.
        - Integrated with a mock external accounting system via HTTP API for billing.
        - Utilized a PDF generation library to create invoices.
        - Dispatched PDF invoices to customers via email.
    - Handled situations where one step from the multi-step orchestrated process fails gracefully.

## Details

- **Technologies Used**: GoLang, PostgreSQL, RabbitMQ, HTTP API, PDF Generation Library.
- **Workflow**:
  - Sign-up: If the contract start date is today, the system adds the task to process and send the invoice to RabbitMQ.
    - Daily Cron Jobs:
      - One for processing invoices billed on that day.
      - Another for retrying failed invoices.
    - Invoice Generation:
      - Internal calculations for invoice amounts.
      - Integration with a mock external accounting system.
      - PDF generation and dispatching invoices to customer emails.

## Repository Structure

- **`project`**: Contains files related to the project setup.
  - **`db`**: Directory for database-related files.
    - **`init.sql`**: SQL script for initializing the database schema.
  - **`docker-compose.yaml`**: Docker Compose configuration file for setting up the project environment.
- **`subscription-backend.code-workspace`**: Visual Studio Code workspace file for the backend service.
- **`subscription-service`**: Main directory for the subscription service.
  - *(Detailed directory structure explanation below)*
- **`README.md`**: Project documentation providing an overview, setup instructions, and usage guidelines.

### Subscription Service Directory Structure

- **`assets`**: Directory for storing assets like logos and templates.
  - **`logo`**: Directory for storing the company logo.
  - **`templates`**: Directory for email templates.
- **`cmd`**: Command directory containing the main executable.
  - **`api`**: Directory for API-related files.
    - **`main.go`**: Entry point for the API server.
- **`internal`**: Directory for internal packages.
  - *(Detailed explanation in the previous section)*
- **`Makefile`**: Makefile for building and running the application.
- **`platforms`**: Directory for platform-specific files.
  - **`routers`**: Directory for router-related files.
    - **`routers.go`**: File defining router setup.
- **`subscription-service.dockerfile`**: Dockerfile for building the subscription service Docker image.
- **`temp`**: Temporary directory (can be removed if not used).

This structure follows best practices for organizing a GoLang backend service, with clear separation of concerns, reusable components, and a modular design. Each directory and file serves a specific purpose, making it easy to navigate and maintain the codebase.

### Usage

1. Clone the repository.
2. Sign up new customers.
3. Ensure the contract start date is correctly recorded.
4. Daily cron jobs will handle invoicing and retrying failed invoices.
5. Monitor system logs and notifications for any errors or exceptions.

### Running the Project

To run the project, follow these steps:

1. Clone the repository:

   ```bash
   git clone <repository_url>

2. Change directory to the project folder:

    ```bash
    cd project

3. Run Docker Compose to build and start the project containers:

    ```bash
    docker-compose up --build -d

4. Download the Postman collection from the provided link.

Use the Postman collection to perform the following actions:

Create a subscription plan.
Sign up a new user.
These steps will set up the project environment and allow you to interact with the subscription service using Postman.

## Monitoring the Project

## MailHog

MailHog is a tool for testing email interactions during development. It captures outgoing emails and displays them in a web interface, allowing you to inspect the content, headers, and other details without actually sending emails to real recipients.

To access MailHog:

1. Open your web browser.
2. Navigate to [http://localhost:8025](http://localhost:8025).
3. You will be directed to the MailHog web interface, where you can view and manage the captured emails.

## Adminer

Adminer is a lightweight database management tool written in PHP. It provides a user-friendly interface for managing databases, executing queries, and performing other administrative tasks.

To access Adminer:

1. Open your web browser.
2. Navigate to [http://localhost:5053](http://localhost:5053).
3. You will be directed to the Adminer login page.
4. Enter the database credentials (username, password, and database name) to log in.
5. Once logged in, you can perform various database management tasks using the Adminer interface.
