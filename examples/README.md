# NCore Examples

Comprehensive, production-ready examples demonstrating all NCore features and patterns. Each example is self-contained
with complete code, documentation, and real-world patterns from actual NCore-based applications.

## Overview

These examples are designed to teach NCore from basics to advanced patterns, with each example building on concepts from
previous ones. They incorporate real patterns from production applications.

## Examples Index

| #                             | Name              | Focus               | Difficulty   | Key Features                       |
|-------------------------------|-------------------|---------------------|--------------|------------------------------------|
| [01](./01-basic-rest-api)     | Basic REST API    | Wire + PostgreSQL   | Beginner     | Wire DI, Ent ORM, CRUD operations  |
| [02](./02-mongodb-api)        | MongoDB API       | Manual DI + MongoDB | Beginner     | No Wire, MongoDB, Manual DI        |
| [03](./03-multi-module)       | Multi-Module App  | Extension System    | Intermediate | Extension Manager, Wrapper Pattern |
| [04](./04-realtime-websocket) | WebSocket Server  | Real-time           | Intermediate | WebSocket Hub, Broadcasting        |
| [05](./05-background-jobs)    | Background Jobs   | Worker Pools        | Intermediate | Job Queue, Status Tracking         |
| [06](./06-event-driven)       | Event-Driven      | Pub/Sub             | Advanced     | Event Bus, Event Sourcing          |
| [07](./07-authentication)     | Authentication    | Security            | Advanced     | JWT, RBAC, Middleware              |
| [08](./08-full-application)   | Full Application  | Everything          | Expert       | Multi-module, Realtime, Jobs, Auth |
| [09](./09-wire)               | Wire ProviderSets | ProviderSets        | Beginner     | ProviderSets, JWT, Worker Pools    |

## Learning Path

### Beginner Path

1. **01-basic-rest-api**: Learn Wire dependency injection and basic CRUD
2. **02-mongodb-api**: Understand manual DI and MongoDB integration

### Intermediate Path

1. **03-multi-module**: Master the extension system and inter-module communication
2. **04-realtime-websocket**: Implement real-time features
3. **05-background-jobs**: Add async processing capabilities

### Advanced Path

1. **06-event-driven**: Build event-driven architectures
2. **07-authentication**: Secure your applications
3. **08-full-application**: Combine everything into a production-ready app

### Supplemental Path

1. **09-wire**: ProviderSets and multi-target wiring

## Feature Matrix

| Feature               | 01  | 02  | 03  | 04  | 05  | 06  | 07  | 08  | 09  |
|-----------------------|-----|-----|-----|-----|-----|-----|-----|-----|-----|
| **Wire DI**           | Yes | No  | No  | No  | No  | No  | Yes | Yes | Yes |
| **Manual DI**         | No  | Yes | Yes | Yes | Yes | Yes | No  | Yes | No  |
| **PostgreSQL**        | Yes | No  | No  | No  | No  | Yes | Yes | Yes | No  |
| **MongoDB**           | No  | Yes | Yes | No  | No  | No  | No  | Yes | No  |
| **Redis**             | No  | No  | No  | Yes | No  | No  | No  | Yes | No  |
| **Ent ORM**           | Yes | No  | No  | No  | No  | No  | No  | Yes | No  |
| **Extension System**  | No  | No  | Yes | No  | No  | No  | No  | Yes | Yes |
| **WebSocket**         | No  | No  | No  | Yes | No  | No  | No  | Yes | No  |
| **Worker Pools**      | No  | No  | No  | No  | Yes | No  | No  | Yes | Yes |
| **Event Bus**         | No  | No  | No  | No  | No  | Yes | No  | Yes | No  |
| **JWT Auth**          | No  | No  | No  | No  | No  | No  | Yes | Yes | Yes |
| **RBAC**              | No  | No  | No  | No  | No  | No  | Yes | Yes | No  |
| **Logging**           | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| **Config Management** | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |

## NCore Modules Used

| Module               | Examples Using It          |
|----------------------|----------------------------|
| `config`             | All                        |
| `logging`            | All                        |
| `data`               | 01, 02, 03, 06, 07, 08, 09 |
| `extension`          | 03, 08, 09                 |
| `security/jwt`       | 07, 08, 09                 |
| `concurrency/worker` | 05, 08, 09                 |
| `messaging`          | 08                         |
| `net`                | All (HTTP responses)       |

## Patterns Demonstrated

### Dependency Injection

- **Wire-based** (Example 01, 08): Automatic DI code generation
- **Manual** (Example 02, 03): Constructor injection pattern

### Architecture Patterns

- **Clean Architecture**: All examples use Handler → Service → Repository
- **Extension Architecture**: Examples 03, 08
- **Event-Driven**: Examples 06, 08
- **Multi-Tenant**: Example 08

### Communication Patterns

- **Direct HTTP**: Examples 01, 02
- **Cross-Module (Wrapper)**: Examples 03, 08
- **Event Bus**: Examples 06, 08
- **WebSocket**: Examples 04, 08

### Data Patterns

- **Repository Pattern**: All examples
- **Unit of Work**: Examples 03, 08
- **Event Sourcing**: Example 06
