# API Gateway [![Lint](https://github.com/TheUnitedCoders/devpost-auth0-api-gateway/actions/workflows/lint.yml/badge.svg)](https://github.com/TheUnitedCoders/devpost-auth0-api-gateway/actions/workflows/lint.yml) [![Test](https://github.com/TheUnitedCoders/devpost-auth0-api-gateway/actions/workflows/test.yml/badge.svg)](https://github.com/TheUnitedCoders/devpost-auth0-api-gateway/actions/workflows/test.yml) [![Build](https://github.com/TheUnitedCoders/devpost-auth0-api-gateway/actions/workflows/build.yml/badge.svg)](https://github.com/TheUnitedCoders/devpost-auth0-api-gateway/actions/workflows/build.yml)

A modern API Gateway ensuring security, scalability, and integration through advanced Auth0 features.

Created as part of participation in the [API World & CloudX 2024](https://api-world-2024-hackathon.devpost.com) Hackathon.

## Solution Description

### Deep Integration with Auth0 and Unique Solution Benefits

1. Exceptional Automation:
With a built-in gRPC handler for retrieving service descriptions, API Gateway configuration updates occur automatically. This eliminates the need for manual configuration changes and reduces the risk of errors.

2. Development Optimization:
Built-in token and user data verification at the API Gateway level reduces the complexity and workload of developing microservices. Developers do not need to duplicate authentication code, making the process faster and more secure.

3. Flexible Access Management (RBAC):
Capability for permission checks at the gateway level allows services to set precise access rights and ensure data protection.

### Key Advantages for Business and Security

1. Enterprise-Level Security:
Our API Gateway offers automatic token verification and M2M (machine-to-machine) authentication integration, ensuring data reliability and security during interactions with various systems and services. This assures downstream services that requests are coming directly from the API Gateway, enhancing trust and security.

2. Scalability Support:
With built-in rate limiting features, the API Gateway helps companies manage load and efficiently allocate resources. The rate limiter enforces load restrictions on downstream services, ensuring high availability and resilience even during periods of increased user traffic.

3. Quick Integration and Deployment:
Developers can leverage a ready-to-use Go SDK with detailed documentation and code examples, significantly reducing the time needed to connect services. This streamlines the integration process, reduces the workload on technical teams, and allows for quicker transitions from development to testing and deployment of new features.

### Alignment with Hackathon Goals, Including Additional Bonus Criteria implementation

1. Security Compliance: Implementation of OAuth 2.0, JWT parsing, and RBAC at the API Gateway level.
2. API Design and Innovation: Unique approach to service description and updating via gRPC, with support for automatic configuration.
3. M2M Authorization: Full support for M2M flow, meeting the hackathon criteria for leveraging Auth0 capabilities.
4. Auditing: Comprehensive logging of all calls with user data inclusion and adaptive protection support.
5. Rate Limiter: Implemented rate limiting configuration at the API Gateway level, aiding effective load management and protecting against request limit attacks.
6. CI/CD: Automated CI/CD processes set up for Docker image builds, testing, and linting.
7. Metrics: Integration with Prometheus for full system state monitoring.

## Installation and usage guide

### Configuring the API Gateway

Create a configuration file config.json specifying all the necessary parameters for launching the API Gateway

```json
{
    "public_listen_address": ":7070", // Port for public request (Default: “:7070”)
    "admin_listen_address": ":7071", // Port for administrative requests (Default: “:7071”)
    "auth0_domain": "<DOMAIN>", // Auth0 domain
    "auth0_audience": "<GATEWAY_AUDIENCE>", // Auth0 audience
    "auth0_client_id": "<CLIENT_ID>", // Auth0 client ID
    "auth0_client_secret": "<CLIENT_SECRET>", // Auth0 client secret
    "redis_address": "localhost:6379", // Address of the Redis server (Default: “localhost:6379”)
    "redis_password": "", // Password for Redis (empty by default)
    "description_sync_period": "1m", // Period for service description updates (Default: “1m”)
    "services": [
        {
            "name": "greeting", // Name of your service
            "m2m_audience": "<AUTH0_AUDIENCE>", // M2M audience for the service
            "address": "127.0.0.1:8001", // Address for service requests
            "timeout": "1m" // Timeout for service requests (Default: “1m”)
        }
    ]
}
```

Please note that JSON format does not support comments, so any lines starting with `//` are only meant as hints to explain each field. Be sure to remove these comments before using the configuration file to avoid errors.

### Launching Redis to Support the API Gateway

To work with rate limiting you need to start Redis. The command to launch Redis using Docker:
```shell
docker run -p 6379:6379 redis:7
```

### Run the API Gateway using Docker with the configuration file config.json:
```bash
docker run -v $(pwd)/config.json:/config.json ghcr.io/theunitedcoders/devpost-auth0-api-gateway:latest
```

### Integrating Microservices

Congrats on the installation! Next, you need to create a new microservice that will be integrated with the API Gateway. To do this, [use our SDK for Go](pkg/sdk/README.md).
