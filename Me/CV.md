
### Languages 

- Go
- Python
- C++
- JavaScript & React
### Skills  

- RESTfull API
- JSON RPC 
- gRPC & Protocol Buffers 
- Microservices 
- Docker 
- GitHub Actions & Gitlab CI
- Bash Scripting
- CMake & make
- PostgreSQL | MySQL | MongoDB | MinIO S3
- Typesense (fast search)
- Payment services Integration 
- Mail service integration (SMTP & SMTPS)
- Phone service integration (SMPP & HTTP)
- Basic AUTH, OAuth 2.0, JWT Auth, API keys, TLS handshake 
- Clean architecture, DDD, Hexagonal Architecture

### Experience 

### First Job: Udevs Outsourcing Company - Software Engineer

At Udevs, my first role as a Software Engineer, I learned how to properly structure code in Golang using a three-tier architecture—presentation (user interface), application (business logic), and data (data storage)—to create scalable projects. I gained experience in:

- Writing REST API methods using the Gin framework.
- Implementing CRUD operations with the repository pattern using the SQLx library.
- Working with PostgreSQL database functions, including optimizing a complex business logic function (reducing execution time from 5-6 seconds to 400ms-1s).
- Managing database migrations with Go-Migrate.

---

### Second Job: IMAN Invest Halal Investment Company - Software Engineer

At this fintech company, I gained my first experience with a fully microservices-based architecture and worked with gRPC and Protocol Buffers. Key achievements included:

- Resolving a critical, long-standing issue within three months that had previously disrupted user balances, requiring manual database corrections and temporary hacks. I debugged the issue across multiple microservices and implemented a permanent fix.
- Actively contributing to a team that rewrote the entire project from scratch, including implementing several microservices and API gateways.
- integrated event driven architecture using Apache Kafka. 
- Learning and implementing authentication methods: JWT, RBAC, Basic Auth, and phone verification.
- Developing REST APIs with the Go-Chi framework.
- Integrating payment providers for local banks and Visa/Mastercard.
- Creating Postman documentation and generating Swagger docs in Golang.

---

### Third Job: PointAI (Startup) - Founding Engineer

As a Founding Engineer at PointAI, an AI-powered customer support service startup, I helped companies deliver high-quality support, reduce wait times, and enhance user satisfaction. Key contributions included:

- Designing a mono-microservices architecture using the Domain-Driven Design (DDD) pattern, blending various programming patterns for fast, clear, and scalable development.
- Integrating external services such as Facebook API, WhatsApp API, Telegram API, Zendesk API, and Intercom API.
- Gaining experience in Prompt Engineering and integrating Large Language Models (LLMs) to create AI-powered chatbots using WebSockets.
- Building the server with the Fiber framework and the repository layer with Pgx and Squirrel SQL builder.
- Handling primary deployment, configuring an Nginx server to serve a static website over HTTPS and proxy requests to the backend.
- Developing a custom library over the OpenAI API to simplify ChatGPT integration, as existing libraries lacked support for the newly released "function call" feature at the time.
- Working with MongoDB.
- Using Python for simple database operation scripts and small FastAPI servers for Python-specific tasks.

---

### Fourth Job: Zakupki.AI - Leading Engineer

As a Leading Engineer at Zakupki.AI, I focused on optimizing systems and implementing innovative solutions:

- Implemented a generic repository pattern using GORM.
- Improved full-text search performance from 20-30 seconds to 200-500ms by integrating Typesense.
- Enhanced filtering performance from 30-40 seconds to 1-2 seconds by optimizing fetching algorithms, SQL queries, and adding column indexing and PostgreSQL tsvector columns where needed.
- Built a multistage, concurrent tender XML file parser in Golang.
- Developed an AI-powered chat system over tender documents by:
    - Loading document paragraphs into Typesense with their vectors.
    - Searching the top 32 relevant paragraphs using the user’s question vector, ranking them with a Cross Encoder, and generating answers from the top 10 paragraphs.
- Created a document converter microservice over gRPC (supporting RTF, PDF, DOC, DOCX, XLS, XLSX, CSV, etc., to HTML).
- Developed a utility microservice in Python over gRPC for ZIP filename encoding and converting formula images to LaTeX.
- Built a Data Delivery Service in C++ using the Crow framework to communicate with government SOAP web services secured by TLS and GOST CryptoPro algorithms, leveraging OpenSSL GOST Engine and cURL. This acted as a bridge/adapter since Golang struggled with GOST CryptoPro certificates.
- Developed a Telegram Bot using the Telebot library and integrated a MiniApp with Go templates, HTML, CSS, and JavaScript.

---

### Fifth Job: EssayAI - Software Engineer

I joined a team of five to build EssayAI, a service for AI-assisted text processing in Russian, with LLMs trained for specific tasks. Key contributions included:

- Implementing a robust subscription system integrated with payment providers.
- Setting up daily cron jobs to renew user subscriptions.
- Configuring Docker containerization and CI/CD pipelines in Gitlab CI and GitHub Actions.
- Experience with mailing service over SMTP/SMTPS.
- Building a monitoring system with Grafana, Prometheus, Loki,  Cronitor, plus system alerts using Logrus library hooks based on error levels.

The project succeeded, achieving over 1,000 subscribers.